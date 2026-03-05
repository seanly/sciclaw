# PHI Mode: One-Click Local Model Flow (Draft)

This defines the expected user experience for `sciclaw app` when a user selects:

`Download local model`

## Product goal

A non-technical user should be able to click once and get a working private local model without touching CLI or config files.

## Runtime states

1. Detect hardware profile (OS, CPU arch, RAM, GPU vendor, GPU VRAM).
2. Resolve profile in `config/phi_model_catalog.json`.
3. Pick backend (`mlx` first on Apple Silicon; `ollama` otherwise unless overridden).
4. Pick model preset (`balanced` by default).
5. Verify backend install.
6. Pull model artifact.
7. Run warmup check.
8. Persist config and switch active mode to PHI.
9. Show success panel with active backend/model and fallback hint.

## UX contract

- Never silently switch to cloud mode.
- If install/pull fails, keep user in PHI setup with a plain-language fix.
- Always show what backend/model was selected and why (hardware profile + preset).
- Keep the default path one-click; advanced controls stay optional.

## Failure handling

- Backend missing:
  - Offer one action: install backend.
- Model pull fails:
  - Retry once.
  - Offer next lower model in profile `allow_models`.
- Warmup fails:
  - Roll to next model in `allow_models`.
  - If all fail, offer explicit cloud fallback button.

## Config writes after success

- `agents.defaults.mode = "phi"` (compat alias to local mode if needed)
- `agents.defaults.local_backend = "<mlx|ollama>"`
- `agents.defaults.local_model = "<model-id>"`
- `agents.defaults.local_preset = "<speed|balanced|quality>"`

## Suggested telemetry (local-only opt-in)

- `phi_setup_started`
- `phi_setup_backend_installed`
- `phi_setup_model_pulled`
- `phi_setup_ready`
- `phi_setup_failed`

Telemetry should avoid prompt content and sensitive user data.
