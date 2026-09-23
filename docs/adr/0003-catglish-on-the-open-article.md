# Catglish click-to-lookup on the open Article

The Reader does not translate. It shows whether Catglish is installed and signed
in. When both are true, learning mode runs on the Article open in the Reading
pane, and a click opens Catglish's word popover. RSSX never looks a word up.

## Status

accepted

## Context

The Reader (`/`) is the home screen: Feed column, Article column, Reading pane.
The Reading pane starts empty. An Article's HTML arrives later, when one is
selected, and Vue replaces it when the selection changes.

Catglish (the ENX Chrome extension) already does click-to-lookup, the word
popover, and learning mode. Its content script is injected on
`https://rssx-lab.wiloon.com/*`, which is the host that serves the Reader today.
Learning mode stays off until something sends `enxRun`. The default site adapter
treats the page as static: it processes whatever text is in the DOM at that
moment and does not watch for a later replacement.

ENX already has one web-to-extension channel, `externally_connectable`, and it
is limited to enx-ui origins (enx adr-019). A page outside that list cannot call
`chrome.runtime.sendMessage`. The extension can still stamp the page, because
the content script shares the DOM.

## Decision

RSSX only displays Catglish presence on the Reader: not installed, installed but
signed out, or ready. It does not implement lookup, and it does not send a
message to the extension.

The extension owns the rest, on this host only:

- It stamps presence onto the page, including whether Catglish is signed in, so
  the Reader can render the three states with no message channel.
- It enables learning mode itself when an Article is open in the Reading pane,
  and runs it again when that Article is replaced.
- The processed region is that Article's title and Feed-provided body. The Feed
  column and the Article column are out of scope.
- Signed-out Catglish does not enable learning mode. The Reader shows that state
  instead of a silent failure on click.

`externally_connectable` stays limited to enx-ui. The extension half of this
decision is recorded in enx `docs/architecture/adr-033-rssx-reading-pane-learning-mode.md`.

## Considered options

- **Ping the extension, as the enx-ui Reader does.** That channel is an explicit
  allowlist of enx-ui origins. Putting the RSSX origin on it would let this app
  drive the extension, which adr-019 kept off the table. A ping at page load
  also fires while the Reading pane is still empty.
- **Enable learning mode when the Reader mounts.** There is no Article yet. A
  later selection replaces the body, and a static pass does not see the new
  HTML.
- **Build the popover in RSSX.** Duplicates click-to-lookup, the dictionary,
  and the sign-in gate that already live in Catglish.

## Consequences

- The content-script match has to be the origin that actually serves the
  Reader. Today that is `rssx-lab.wiloon.com`. A host change has to update the
  match in the same change.
- Switching Articles must re-run learning mode. Otherwise the new body has no
  click-to-lookup, or Vue's render wipes the previous pass.
- Lookup still requires a signed-in Catglish session and still counts against
  Catglish's lookup quota. RSSX does not grow a dictionary or a billing path.
