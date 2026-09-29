package legacy

import (
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"

	"rssx/common"
	"rssx/user"
)

var miniRedis *miniredis.Miniredis

func TestMain(m *testing.M) {
	mr, err := miniredis.Run()
	if err != nil {
		panic("failed to start miniredis: " + err.Error())
	}
	miniRedis = mr
	// Must be set before the first redisx call (the pool initialises once).
	os.Setenv("REDIS_ADDRESS", mr.Addr())
	code := m.Run()
	mr.Close()
	os.Exit(code)
}

func reset(t *testing.T) {
	t.Helper()
	common.InitForTesting()
	miniRedis.FlushAll()
}

func addUser(t *testing.T, id, name string) {
	t.Helper()
	if err := common.DB.Create(&common.User{Id: id, Name: name, Password: "x"}).Error; err != nil {
		t.Fatalf("create user %s: %v", name, err)
	}
}

func subscribe(t *testing.T, userId string, feedIds ...int64) {
	t.Helper()
	for _, fid := range feedIds {
		if err := common.DB.Create(&common.UserFeed{UserId: userId, FeedId: fid}).Error; err != nil {
			t.Fatalf("subscribe %s to %d: %v", userId, fid, err)
		}
	}
}

func subscriptions(t *testing.T, userId string) map[int64]bool {
	t.Helper()
	var rows []common.UserFeed
	common.DB.Where("user_id = ?", userId).Find(&rows)
	out := map[int64]bool{}
	for _, r := range rows {
		out[r.FeedId] = true
	}
	return out
}

func TestMigrate_SoleUserInheritsLegacyData(t *testing.T) {
	reset(t)
	addUser(t, "owner-uuid", "wiloon")
	subscribe(t, user.LegacyUserId, 1, 2, 3)
	miniRedis.Set("read_index:0:1", "12345")
	miniRedis.SAdd("read_mark:0:2", "a", "b")
	miniRedis.Set("read_index:other:1", "999")

	if err := MigrateLegacyUserData(common.DB, ""); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if got := subscriptions(t, "owner-uuid"); len(got) != 3 || !got[1] || !got[2] || !got[3] {
		t.Errorf("owner subscriptions = %v, want feeds 1,2,3", got)
	}
	if got := subscriptions(t, user.LegacyUserId); len(got) != 0 {
		t.Errorf("legacy subscriptions left = %v, want none", got)
	}
	if v, _ := miniRedis.Get("read_index:owner-uuid:1"); v != "12345" {
		t.Errorf("read_index moved value = %q, want 12345", v)
	}
	if members, _ := miniRedis.Members("read_mark:owner-uuid:2"); len(members) != 2 {
		t.Errorf("read_mark moved members = %v, want [a b]", members)
	}
	for _, k := range []string{"read_index:0:1", "read_mark:0:2"} {
		if miniRedis.Exists(k) {
			t.Errorf("legacy key %s still exists", k)
		}
	}
	if v, _ := miniRedis.Get("read_index:other:1"); v != "999" {
		t.Error("another user's key was touched")
	}

	// Idempotent: a second run is a no-op.
	if err := MigrateLegacyUserData(common.DB, ""); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if got := subscriptions(t, "owner-uuid"); len(got) != 3 {
		t.Errorf("after second run owner subscriptions = %v", got)
	}
}

func TestMigrate_OverlappingSubscriptionKeepsOwnerRow(t *testing.T) {
	reset(t)
	addUser(t, "owner-uuid", "wiloon")
	subscribe(t, "owner-uuid", 1)
	subscribe(t, user.LegacyUserId, 1, 2)
	miniRedis.Set("read_index:owner-uuid:1", "newer")
	miniRedis.Set("read_index:0:1", "older")

	if err := MigrateLegacyUserData(common.DB, ""); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var rows int64
	common.DB.Model(&common.UserFeed{}).Where("user_id = ?", "owner-uuid").Count(&rows)
	if rows != 2 {
		t.Errorf("owner rows = %d, want 2 (no duplicate for feed 1)", rows)
	}
	if v, _ := miniRedis.Get("read_index:owner-uuid:1"); v != "newer" {
		t.Errorf("owner's existing read index was overwritten: %q", v)
	}
}

func TestMigrate_AmbiguousOwnerLeavesDataUntilConfigured(t *testing.T) {
	reset(t)
	addUser(t, "u1", "wiloon")
	addUser(t, "u2", "guest")
	subscribe(t, user.LegacyUserId, 1)

	if err := MigrateLegacyUserData(common.DB, ""); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if got := subscriptions(t, user.LegacyUserId); !got[1] {
		t.Fatal("legacy data moved without a decidable owner")
	}

	if err := MigrateLegacyUserData(common.DB, "wiloon"); err != nil {
		t.Fatalf("migrate with owner: %v", err)
	}
	if got := subscriptions(t, "u1"); !got[1] {
		t.Errorf("configured owner did not receive legacy data: %v", got)
	}
	if got := subscriptions(t, "u2"); len(got) != 0 {
		t.Errorf("other user received legacy data: %v", got)
	}
}

func TestMigrate_UnknownConfiguredOwnerDoesNothing(t *testing.T) {
	reset(t)
	addUser(t, "u1", "wiloon")
	subscribe(t, user.LegacyUserId, 1)

	if err := MigrateLegacyUserData(common.DB, "nobody"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if got := subscriptions(t, user.LegacyUserId); !got[1] {
		t.Fatal("legacy data moved to an unknown owner")
	}
}
