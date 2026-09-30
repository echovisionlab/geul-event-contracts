# API query contracts

## Artist works

`ArtistService.GetWorks` returns each work's canonical `api.open.v1.WorkType` in
`ArtistWork.type`. Clients can use it for labels and type filters instead of
assuming every artist work is a music project.

## Form submissions

`FormService.ListFormSubmissions` keeps its form-scoped management permission
check and supports these optional filters:

- `search`: a case-insensitive literal substring of the submitted JSON data.
  `%`, `_`, and `\\` are literal characters; callers do not supply SQL
  wildcards. The value is limited to 200 characters.
- `country_code`: an exact two-letter ASCII country code. Lowercase input and
  surrounding whitespace are normalized to uppercase before matching.
- `created_at_from`: inclusive lower timestamp bound.
- `created_at_before`: exclusive upper timestamp bound.

When both timestamp bounds are present, `created_at_from` must be earlier than
`created_at_before`. `sorts` accepts `createdAt` and `created_at` as aliases for
the submission creation timestamp, with `ASC` or `DESC` order. The default page
size remains 50 and requests are capped at 100. The default sort remains newest
first. All sort modes use submission ID descending as a final tie-breaker so
pagination is stable. The response total is computed from the same filtered
set as the page.
