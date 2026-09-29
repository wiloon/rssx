// Package legacy moves data from the single-user era onto real accounts.
package legacy

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"

	"rssx/common"
	"rssx/storage/redisx"
	"rssx/user"
	log "rssx/utils/logger"
)

// legacyRedisPrefixes are the per-user Redis key families; the user id is the
// segment right after the prefix ("read_index:<userId>:<feedId>").
var legacyRedisPrefixes = []string{"read_index:", "read_mark:"}

// MigrateLegacyUserData hands everything owned by user.LegacyUserId — the
// user_feeds rows and the Redis read state — to a real account. The owner is
// the user named ownerName, or, when ownerName is empty, the only registered
// user. With no legacy data, or no unambiguous owner, it does nothing.
//
// It is idempotent: Redis keys move first and SQL rows last, so a crash in
// between is finished by the next run.
func MigrateLegacyUserData(db *gorm.DB, ownerName string) error {
	var legacyRows int64
	if err := db.Model(&common.UserFeed{}).Where("user_id = ?", user.LegacyUserId).Count(&legacyRows).Error; err != nil {
		return fmt.Errorf("count legacy subscriptions: %w", err)
	}
	if legacyRows == 0 {
		return nil
	}

	owner, err := resolveOwner(db, ownerName)
	if err != nil {
		return err
	}
	if owner == nil {
		log.Warnf("legacy data: %d subscriptions still belong to user %q; set rssx.legacy-owner (env RSSX_LEGACY_OWNER) to the username that should own them",
			legacyRows, user.LegacyUserId)
		return nil
	}

	moved, err := moveRedisKeys(user.LegacyUserId, owner.Id)
	if err != nil {
		return fmt.Errorf("move legacy read state: %w", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// A feed the owner already subscribes to keeps the owner's row.
		if err := tx.Where("user_id = ? AND feed_id IN (?)", user.LegacyUserId,
			tx.Model(&common.UserFeed{}).Select("feed_id").Where("user_id = ?", owner.Id)).
			Delete(&common.UserFeed{}).Error; err != nil {
			return err
		}
		return tx.Model(&common.UserFeed{}).Where("user_id = ?", user.LegacyUserId).
			Update("user_id", owner.Id).Error
	})
	if err != nil {
		return fmt.Errorf("move legacy subscriptions: %w", err)
	}

	log.Infof("legacy data: moved %d subscriptions and %d read-state keys from user %q to %q (%s)",
		legacyRows, moved, user.LegacyUserId, owner.Name, owner.Id)
	return nil
}

// resolveOwner returns the target account, or nil when it cannot be decided.
func resolveOwner(db *gorm.DB, ownerName string) (*common.User, error) {
	if ownerName != "" {
		var owner common.User
		err := db.Where("name = ?", ownerName).First(&owner).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warnf("legacy data: rssx.legacy-owner %q is not a registered user", ownerName)
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("look up legacy owner %q: %w", ownerName, err)
		}
		return &owner, nil
	}

	var users []common.User
	if err := db.Limit(2).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	if len(users) != 1 {
		return nil, nil
	}
	return &users[0], nil
}

// moveRedisKeys renames every per-user key of fromUser to toUser. A key whose
// target already exists is left in place rather than overwriting newer state.
func moveRedisKeys(fromUser, toUser string) (int, error) {
	moved := 0
	for _, prefix := range legacyRedisPrefixes {
		from := prefix + fromUser + ":"
		keys, err := scanKeys(from + "*")
		if err != nil {
			return moved, err
		}
		for _, key := range keys {
			target := prefix + toUser + ":" + strings.TrimPrefix(key, from)
			renamed, err := redis.Bool(redisx.Exec("RENAMENX", key, target))
			if err != nil {
				return moved, fmt.Errorf("rename %s: %w", key, err)
			}
			if !renamed {
				log.Warnf("legacy data: %s already exists, keeping it and leaving %s", target, key)
				continue
			}
			moved++
		}
	}
	return moved, nil
}

func scanKeys(pattern string) ([]string, error) {
	var keys []string
	cursor := "0"
	for {
		reply, err := redis.Values(redisx.Exec("SCAN", cursor, "MATCH", pattern, "COUNT", 500))
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", pattern, err)
		}
		if len(reply) != 2 {
			return nil, fmt.Errorf("scan %s: unexpected reply length %d", pattern, len(reply))
		}
		cursor, err = redis.String(reply[0], nil)
		if err != nil {
			return nil, err
		}
		batch, err := redis.Strings(reply[1], nil)
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		if cursor == "0" {
			return keys, nil
		}
	}
}
