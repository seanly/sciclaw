# sciClaw Next Release Announcement Draft

Use this file when the next release tag is cut. Replace:
- `vNEXT` with the actual version
- `RELEASE_URL` with the real GitHub release link

Keep the tone aligned with prior announcement files:
- short opening line
- concrete feature list
- problem/benefit framing
- one clear install/upgrade path
- no hype language that overclaims

Do not show real PHI, API keys, email addresses, bot tokens, or private channel names in screenshots.

---

## Tweet 1 - Main Release Announcement

sciClaw vNEXT is out.

This release makes local mode more usable and more honest:
- built-in PHI eval now checks real local behavior instead of just “is Ollama installed”
- the app shows whether a machine is good for interactive local work or just a slower fallback
- PDF form workflows now have a typed tool path instead of ad hoc shell usage

RELEASE_URL

---

## Tweet 2 - Local PHI Mode

You do not have to choose between “all cloud” and “all local.”

sciClaw can now make that distinction much clearer:
- route sensitive rooms to local PHI mode
- keep other rooms on normal cloud defaults
- see local suitability directly in the app before trusting a machine for private work

That is a much better operational model than pretending every laptop gives the same local AI experience.

---

## Tweet 3 - Honest Local Diagnostics

One of the hardest problems with local models is not setup. It is knowing whether they are actually usable.

sciClaw now keeps local eval results around and shows:
- last eval time
- backend/model used
- probe status
- fallback behavior
- whether the machine is “good interactive” or just “fallback only”

Especially important for CPU-only boxes.

---

## Tweet 4 - PDF Form Workflow

PDF form filling is now a first-class workflow instead of “hope the agent writes the right shell command.”

The new path is:
- inspect the PDF form
- export the field schema
- fill with a typed tool path
- use the new `acroform-fill` skill as the workflow guide

This is safer and much easier to reason about than free-form shell.

---

## Tweet 5 - Upgrade / Try It

For existing installs:

```bash
brew update && brew upgrade sciclaw
sciclaw doctor
sciclaw app
```

For new installs:

```bash
brew tap drpedapati/tap && brew install sciclaw
sciclaw app
```

Docs: https://sciclaw.dev/docs.html

---

## Isolated Engagement Tweet 1 - Room-Level Local Routing

Hook:

You should be able to keep one room local without turning your whole setup into local mode.

Body:

That is the model we are pushing in sciClaw:
- one sensitive room can run on a local model
- everything else can stay on cloud defaults
- the Routing tab now makes that state much clearer

Screenshot:
- Routing tab
- one mapped room highlighted
- detail pane visible
- synthetic channel name only

---

## Isolated Engagement Tweet 2 - CPU-Only Honesty

Hook:

CPU-only local AI is useful. It is just not the same thing as GPU local AI.

Body:

The worst UX is pretending they are equivalent.

sciClaw now says that directly in the app:
- interactive local on a strong GPU box
- fallback-only local on a slower CPU-only box

That is a better product than vague “local supported” claims.

Screenshot:
- PHI tab on a CPU-only machine
- show `Suitability`, `Eval timings`, and `Last Output`

---

## Isolated Engagement Tweet 3 - Native PDF Tooling

Hook:

If an agent has to invent raw shell for every document workflow, it will eventually do something stupid.

Body:

We started fixing that in sciClaw by giving PDF forms a typed adapter layer:
- inspect
- schema
- fill

Then we layered a focused skill on top.

That is the right split:
- code owns execution safety
- skill owns workflow guidance

Screenshot:
- terminal with a clean `pdf_form_*` tool result
- optional side-by-side output PDF filename in workspace

---

## Screenshot Guidance

### Best release screenshot

Use a side-by-side composite:
- left: PHI tab showing local suitability and eval timing
- right: Routing tab showing one room set to local

Why:
- communicates the local story in one image
- shows both machine-level readiness and room-level control

### Best second screenshot

Use a terminal + artifact pairing:
- terminal pane with PDF form inspect/schema/fill results
- finder/file list showing output files

Why:
- demonstrates the new typed workflow without needing real document contents

### Screenshot rules

- Use synthetic or public data only
- No real PHI
- No real Discord/Telegram private channel names
- No API keys or email addresses
- Prefer short workspace paths
- Keep the visible version number in frame if possible

### Caption patterns that match prior style

- “sciClaw vNEXT is out.”
- “This release makes local mode more usable and more honest.”
- “You do not have to choose between all cloud and all local.”
- “This is safer than free-form shell.”

### Phrases to avoid

- “game changer”
- “revolutionary”
- “perfect”
- “production-ready for every machine”
- “just works” unless the exact scope is narrow and defensible
