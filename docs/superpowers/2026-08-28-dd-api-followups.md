# dd-api follow-ups after `resolved-decode`

What the `resolved-decode` branch does **not** fix, and what gates its deploy.
Written at the final whole-branch review of `28152d9..HEAD` (branch
`resolved-decode`, stacked on `r2-device-definitions` / PR #314).

Plan: `docs/superpowers/plans/2026-08-28-resolved-decode.md`
Ledger: `.superpowers/sdd/2026-08-28-resolved-decode/progress.md`
Contract: `../definitions-worker/schema/{template,resolved}.schema.json`

---

## 1. Deploy blocker: CLEARED on branch `shrink-dd-api`

**Status: fixed.** This section described the blocker; it is resolved by the
`shrink-dd-api` branch stacked on `resolved-decode`. Recorded here rather than
deleted, because the failure mode is worth keeping.

**What it was.** The worker's route table is `/t/:id`, `/schema/:doc`,
`POST /admin/{import,build,build/publish}`, `GET /admin/history`, and a 404 for
everything else. There is no `/definitions/:id` and no `/manifest.json`, and
dd-api called both. `Create()` read `GET /definitions/<id>`, got the catch-all
404, and `fetchDocFrom` translated that to `(nil, nil)` -- "does not exist" --
so the existence guard always passed; it then `PUT` to a route that does not
exist. Loud, not corrupting, but every catalog miss on the decode path and
every search sync would have failed.

**What changed.**

| Was | Now |
|---|---|
| `Create()` → `PUT /definitions/<id>` | `PUT /t/<id>`, a template with one `Base` trim |
| `Delete()` → `DELETE /definitions/<id>` | `DELETE /t/<id>` |
| existence check → `fetchDocFresh` | `GetTemplateByIDFresh`, on the `ErrTemplateNotFound` sentinel |
| `GetDeviceDefinitionByID` → `/definitions/<id>.json` | template read, flattened to shared attributes only |
| `manifest()`, `CatalogIDs`, `PinCatalogSnapshot` | deleted with the search sync |
| `GetDeviceDefinitions` + its two handlers | deleted; both handlers were registered and dispatched by nothing |

dd-api now requests only `/t/…` and `/schema/…`.

**Two decisions inside it.**

- Created templates carry attributes folded onto the DeviceType vocabulary by
  `internal/core/vocabulary`, a port of the extraction's `vocabulary.mjs`. The
  two write into the same contract, so they are held to parity by a test
  against the JavaScript's own output on the real vocabulary. A vocabulary that
  cannot be fetched aborts the create rather than writing fewer attributes.
- The legacy flat shape carries only attributes **shared by every trim**.
  Filling its single slot per attribute from an arbitrary trim is what produced
  the blended record in the first place.

**Cutover ordering this imposes.** The worker must have built and published an
index before dd-api ships without its sync, and the Typesense collection
currently squatting the alias name has to be dropped so the alias can take it.

## 2. Not fixed, recorded

### 2.1 `bulk_validate_vin.go` has a guaranteed panic — pre-existing, in dead code

`internal/core/commands/bulk_validate_vin.go:68` does an unchecked
`devideDefinition.(*coremodels.GetDeviceDefinitionQueryResult)` on the result
of `GetDeviceDefinitionByIDQueryHandler.Handle`, which returns
`*coremodels.Template`. It returned `*DeviceDefinitionTablelandModel` before
this branch, so the assertion was **already wrong on `28152d9`** — this branch
changes which wrong type it panics on, not whether it panics. Line 79 then
indexes `.DeviceStyles[0]` on a slice that could be empty.

`BulkValidateVinCommand` is registered in `api.go:96` but **nothing dispatches
it** — no HTTP route, no gRPC method, no other caller. It is unreachable, and
fiber's `recover.New()` would turn the panic into a 500 if it ever became
reachable. Delete the command, or fix and wire it — do not leave it as a
tripwire for whoever adds the route.

### 2.2 `GetDeviceStyleByID` lost every trim-varying attribute

`get_ds_by_id.go:76` builds `DeviceAttributes` from `dd.Attributes` — the
**template-level** attributes only. For a multi-trim template those are exactly
the attributes every trim agrees on, so `fuel_tank_capacity_gal`, `mpg_city`
and the rest are now absent from this response where the flat definition used
to carry them (blended and often wrong, but present).

The handler has the style's `Name` and could narrow via a `styleName` selector
— but the extraction emits `manufacturerCode` selectors only today (see 2.3),
and the style row carries no manufacturer code, so it could only ever resolve
`model-only`. **Blocked on the worker emitting `styleName` selectors.** When
that lands, run `MatchTrim` here with `MatchSignals{StyleName: ds.Name}` and
source the attributes from the resolved result.

### 2.3 Trim matching reaches drivly-decoded VINs only

Adjudicated in the ledger and unchanged: `manufacturerCode` is populated by
`buildFromDrivly` alone; vincario, autoiso, DATGroup, Japan17VIN, CarVX and
Tesla all leave it empty. Templates key trims on `manufacturerCode` only, so
every non-drivly decode returns `model-only`. That is visible in the response
rather than silent, and now countable —
`device_definitions_api_vin_trim_match_quality_total{quality,source}` was added
by this review for exactly this.

**Watch when the follow-up lands:** whatever `styleName` values the extraction
emits must agree with what dd-api sends. dd-api's drivly style name is
`strings.TrimSpace(trim + " " + subModel)` (`buildDrivlyStyleName`); other
providers build it differently again (`autoiso` → `SubModel`, `DATGroup` →
`SubModelName`). Matching is case-insensitive but otherwise exact, so a
format disagreement produces `model-only` across the board with no error.

### 2.4 `docs/swagger.json` is stale

`GET /device-definitions/{id}` now returns `models.Template`. The handler
annotation was corrected on this branch; `docs/swagger.json` and
`docs/swagger.yaml` still describe `DeviceDefinitionTablelandModel`. They are
generated (`make gen-swag`) and `swag` is not installed in this environment, so
regeneration was left out rather than hand-edited. **Run `make gen-swag` before
merge or as the first commit after it.**

Separately: this is an **undeclared response-shape change on a public REST
endpoint**. Any consumer reading `metadata.device_attributes` from it breaks.
That is a consequence of the migration, not a defect, but it needs announcing.

### 2.5 `POST /device-definitions/decode-vin` surfaces none of the new fields

`device_definition_handler.go`'s `DecodeVINResponse` returns only
`deviceDefinitionId` and `newTransactionHash`. Trim, match quality, candidates,
matched-by and hardware template id reach gRPC consumers only. Pre-existing
shape, but materially more lossy now. Widen it if any REST consumer needs the
trim.

### 2.6 `GetTemplateByIDFresh` has no production caller

Its only caller was `bulk_update_powertrain`, deleted by this branch's own
commit `675c939` along with the legacy `Update()` write path. Kept because it
is the read half of the `Create()` migration in item 1; its comment now says so
explicitly. **If that migration lands without adopting it, delete it** — do not
keep it alive on the strength of a comment.

### 2.7 Minor, inherited

- `get_dd_dynamic_filter.go:141` (`buildDeviceDefinitionQueryResponseFromTemplate`)
  ignores `GetManufacturer`'s error and then dereferences `manufacturer.TokenID`.
  Identity's client can return `(nil, nil)`, so this can nil-panic. It mirrors
  the pre-existing `buildDeviceDefinitionQueryResponse` at line 102 exactly —
  fix both together or neither.
- `MatchTrim` has no nil-`tmpl` guard. Both call sites check; note it if a
  third is added.
- The attribute merge in `MatchTrim` is a shallow copy, so a mutable reference
  value would alias the template. Unreachable while the contract restricts
  attribute values to string/number/integer/boolean.
- `decode_vin.go` still creates a definition when `GetTemplateByID` returns a
  nil template with a nil error. Unreachable through the real gateway —
  `fetchTemplateDocFrom` returns either an error or a non-nil template — so it
  is defensive only.

---

## 3. What is *not* a follow-up — verified holding

Recorded so a later reader does not have to re-derive them.

- **`processDeviceStyle` still receives the independently-derived powertrain.**
  `pt` is assigned only at `decode_vin.go:257` and `:260`, both from
  `powerTrainTypeService`, and read at `:263` and at the `processDeviceStyle`
  call. The resolved-trim block resets `resp.Powertrain` and never touches
  `pt`. Resolved-trim data does not reach `device_styles`, so today's decode
  cannot shape tomorrow's templates through the extraction pipeline's input.
- **Proto changes are additive.** Fields 1–15 are byte-identical; 16 and 17 are
  appended. Regenerated with the pinned `protoc-gen-go` v1.30.0.
- **No fallback to `definitions/<id>.json` on the template read.** Asserted by
  a test that pins the exact request path, on both `GetTemplateByID` and
  `GetTemplateByIDFresh`.
