# Wildlens

The wildlife field journal uses the existing Go, templ and SQLite stack.

Run `make templ` after changing templates, then `make web`. The default URL is
http://localhost:8080. To use a different port, run `go run . web -p 8084`.

| Route | Purpose |
| --- | --- |
| `/` | Overview with collection totals, recent encounters, featured animals and an animated globe |
| `/sightings` | Searchable sightings, animal/country filters, sorting and pagination |
| `/sightings/{id}` | Observation details, historical weather, location and credited photographs |
| `/gallery` | Clickable animal portraits |
| `/animals/{gbifKey}` | Animal facts with sources, taxonomy, observation map and recent sightings |
| `/map` | Interactive map and a paginated text/card alternative |
| `/statistics` | Monthly activity, animal/country rankings and collection summary |

Collection views accept `animal` (GBIF species key), `country` (two-letter code),
`q` (search), `sort` (`newest`, `oldest`, or `animal`), and `page`. Filters are
combined and reports always cover all matching observations, not just one page.
All sightings are read for each collection request; this suits the workshop
collection. A larger deployment should move aggregation and pagination into SQL.

Weather is the stored reading at the time of an observation; it is never presented
as a current forecast. Missing weather, photos, results and records have explicit
fallbacks. Unknown routes and records return a custom page with HTTP status 404.

Leaflet 1.9.4 is served locally from `assets/vendor/leaflet`, including its licence.
Map tiles come from OpenStreetMap and observation photos from their original
providers; these need internet access. Maps retain their record links and card
alternatives if tiles fail. Fonts use Google Fonts with local system fallbacks.

The globe uses Canvas 2D to project a rotating 3D sphere, without a WebGL dependency.
Its outlines and accent dots are decorative; `/map` plots the actual records.
Pointer effects run only on devices with a fine pointer. Animations respect
`prefers-reduced-motion`, pause when offscreen or in a hidden tab, and can be paused
with the sparkle button in the top bar. This choice is remembered locally.

Animal facts live in `views/pages/facts.go` with direct source links. Species
without a curated fact show taxonomy context from the stored GBIF record.
Photographer credits, licence links and original images are available on detail
pages. No database migration or API key is required.

Validation: `go test ./...`; focused view and handler checks: `go test ./web/...`.
