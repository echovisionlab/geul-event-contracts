# RPC Usage Audit

> **Superseded and not trusted as consumer evidence.** This document contains
> the first-pass audit, including service-level claims and automated search
> results. It must not be used to retain an RPC. The manual per-procedure audit
> is tracked in `rpc-inventory.json` under `consumerAuditStatus` and each
> procedure's explicit `consumers` records.

The protobuf service declarations are the only RPC contract source. This audit
records direct production call chains and executable test evidence; it is not a
second endpoint allowlist.

An RPC is retained only when a production call chain reaches it and its API
implementation is exercised by integration tests. A declaration that only
returns a retirement or compatibility error is removed from the proto together
with its generated client/server surface and API method.

## `api.intra.v1.InternalEmailTemplateService`

| RPC                       | Production call chain                                                                                                                                                                                                                                       | Policy boundary                                                     | Executable evidence                                                             | Decision                          |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `SaveDocument`            | None. The API method always returns `FailedPrecondition` for the retired shared document.                                                                                                                                                                   | Explicitly denied by Oathkeeper's `/api.intra.v1.*` rule.           | API retirement unit test only.                                                  | Remove.                           |
| `LoadDocument`            | None. The API method always returns `FailedPrecondition` for the retired shared document.                                                                                                                                                                   | Explicitly denied by Oathkeeper's `/api.intra.v1.*` rule.           | API retirement unit test only.                                                  | Remove.                           |
| `SaveTranslationDocument` | `editor-collab/handlers/email-template.ts:emailTemplateHandler.store` -> `lib/api-client.ts:saveEmailTemplateTranslationDocument` -> full Connect procedure. The handler is selected by `handlers/index.ts` and executed by `persistCollaborativeDocument`. | Explicitly denied by Oathkeeper; collab calls the backend directly. | Collab handler and API-client tests exist. API integration coverage is missing. | Retain; add API integration test. |
| `LoadTranslationDocument` | `editor-collab/handlers/email-template.ts:emailTemplateHandler.load` -> `lib/api-client.ts:loadEmailTemplateTranslationDocument` -> full Connect procedure. The handler is selected by `handlers/index.ts` and executed by Hocuspocus `onLoadDocument`.     | Explicitly denied by Oathkeeper; collab calls the backend directly. | Collab handler and API-client tests exist. API integration coverage is missing. | Retain; add API integration test. |

## `api.intra.v1.InternalEmailLayoutService`

| RPC                       | Production call chain                                                                                                                                                                                                                                 | Policy boundary                                                     | Executable evidence                                                             | Decision                          |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `SaveDocument`            | None. The API method always returns `FailedPrecondition` for the retired shared document.                                                                                                                                                             | Explicitly denied by Oathkeeper's `/api.intra.v1.*` rule.           | API retirement unit test only.                                                  | Remove.                           |
| `LoadDocument`            | None. The API method always returns `FailedPrecondition` for the retired shared document.                                                                                                                                                             | Explicitly denied by Oathkeeper's `/api.intra.v1.*` rule.           | API retirement unit test only.                                                  | Remove.                           |
| `SaveTranslationDocument` | `editor-collab/handlers/email-layout.ts:emailLayoutHandler.store` -> `lib/api-client.ts:saveEmailLayoutTranslationDocument` -> full Connect procedure. The handler is selected by `handlers/index.ts` and executed by `persistCollaborativeDocument`. | Explicitly denied by Oathkeeper; collab calls the backend directly. | Collab handler and API-client tests exist. API integration coverage is missing. | Retain; add API integration test. |
| `LoadTranslationDocument` | `editor-collab/handlers/email-layout.ts:emailLayoutHandler.load` -> `lib/api-client.ts:loadEmailLayoutTranslationDocument` -> full Connect procedure. The handler is selected by `handlers/index.ts` and executed by Hocuspocus `onLoadDocument`.     | Explicitly denied by Oathkeeper; collab calls the backend directly. | Collab handler and API-client tests exist. API integration coverage is missing. | Retain; add API integration test. |

## `api.intra.v1.InternalGatewayAuthorizationService`

| RPC                      | Production call chain                                                                                                          | Decision                                                                                                            |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------- |
| `AuthorizeGatewayAccess` | Oathkeeper author/admin `remote_json` rules call the API's private Connect HTTP procedure after Kratos session authentication. | Retain; Oathkeeper is the direct production consumer and the identity integration contract exercises this boundary. |

All remaining `api.intra.v1` services below share the same network boundary:
Oathkeeper explicitly denies `/api.intra.v1.*`; trusted services call the backend
directly.

## `api.intra.v1.InternalArtistService`

| RPC                       | Production call chain                                                                                                                                                  | Decision                               |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| `SaveDocument`            | `editor-collab/handlers/artist.ts:artistHandler.store` -> `saveArtistDocument`; the handler is registered in `handlers/index.ts` and run by collaborative persistence. | Retain; API integration audit pending. |
| `LoadDocument`            | `artistHandler.load` -> `loadArtistDocument`; Hocuspocus `onLoadDocument` selects the registered handler.                                                              | Retain; API integration audit pending. |
| `SaveTranslationDocument` | Locale branch of `artistHandler.store` -> `saveArtistTranslationDocument`.                                                                                             | Retain; API integration audit pending. |
| `LoadTranslationDocument` | Locale branch of `artistHandler.load` -> `loadArtistTranslationDocument`.                                                                                              | Retain; API integration audit pending. |

## `api.intra.v1.InternalCampaignService`

| RPC                       | Production call chain                                                                 | Decision                               |
| ------------------------- | ------------------------------------------------------------------------------------- | -------------------------------------- |
| `SaveDocument`            | `editor-collab/handlers/campaign.ts:campaignHandler.store` -> `saveCampaignDocument`. | Retain; API integration audit pending. |
| `LoadDocument`            | `campaignHandler.load` -> `loadCampaignDocument`.                                     | Retain; API integration audit pending. |
| `SaveTranslationDocument` | Locale branch of `campaignHandler.store` -> `saveCampaignTranslationDocument`.        | Retain; API integration audit pending. |
| `LoadTranslationDocument` | Locale branch of `campaignHandler.load` -> `loadCampaignTranslationDocument`.         | Retain; API integration audit pending. |

## `api.intra.v1.EmailCourierService`

| RPC         | Production call chain                                                                                                                                                | Decision                                      |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| `SendEmail` | Kratos HTTP courier uses `identity/compose/identity.yml:COURIER_HTTP_REQUEST_CONFIG_URL` with the exact full procedure and reaches the API courier handler directly. | Retain; API/Kratos integration audit pending. |

## `api.intra.v1.InternalFormService`

| RPC            | Production call chain                                                                                                                                         | Decision                               |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| `SaveDocument` | `editor-collab/handlers/form.ts:formHandler.store` -> `saveFormDocument`; sends canonical title/schema, exact-locale presence, contributor, and revision CAS. | Retain; API integration audit pending. |
| `LoadDocument` | `formHandler.load` -> `loadFormDocument`; returns source and exact requested-locale canonical values, row existence, presence, and revisions.                 | Retain; API integration audit pending. |

## `api.intra.v1.InternalLabelService`

| RPC                       | Production call chain                                                        | Decision                               |
| ------------------------- | ---------------------------------------------------------------------------- | -------------------------------------- |
| `SaveDocument`            | `editor-collab/handlers/label.ts:labelHandler.store` -> `saveLabelDocument`. | Retain; API integration audit pending. |
| `LoadDocument`            | `labelHandler.load` -> `loadLabelDocument`.                                  | Retain; API integration audit pending. |
| `SaveTranslationDocument` | Locale branch of `labelHandler.store` -> `saveLabelTranslationDocument`.     | Retain; API integration audit pending. |
| `LoadTranslationDocument` | Locale branch of `labelHandler.load` -> `loadLabelTranslationDocument`.      | Retain; API integration audit pending. |

## `api.intra.v1.InternalMapService`

| RPC                    | Production call chain                                                     | Decision                                             |
| ---------------------- | ------------------------------------------------------------------------- | ---------------------------------------------------- |
| `SaveMapThemeSnapshot` | `mapThemeHandler.store` -> typed snapshot CAS via `saveMapThemeSnapshot`. | Retain; DB structured fields are durable authority.  |
| `LoadMapThemeSnapshot` | `mapThemeHandler.load` -> typed snapshot via `loadMapThemeSnapshot`.      | Retain; Yjs is rebuilt only for transient live sync. |

## `api.intra.v1.InternalOgService`

| RPC                    | Production call chain                                                                                      | Decision                                                          |
| ---------------------- | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| `ClaimOgGeneration`    | `og/src/services/processor.ts:createOgProcessor` claims the database generation identity before rendering. | Retain; a non-claimable generation is skipped, never deferred.    |
| `CompleteOgGeneration` | `createOgProcessor` reports the replay-safe object write result after rendering.                           | Retain.                                                           |
| `FailOgGeneration`     | `createOgProcessor` reports a rendering, provider, or lock failure as a terminal lifecycle state.          | Retain; recovery creates a new generation via Admin regeneration. |

## `api.intra.v1.InternalSiteSettingService`

Removed. The claimed generation now carries the immutable rendering snapshot;
the worker no longer reads mutable site settings or keeps an OG config cache.

## `api.intra.v1.InternalPageService`

| RPC                                    | Production call chain                                                                                                                                                                                      | Decision                                                         |
| -------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| `ApplyVersionProjectionBackfill`       | `scripts/ops/backfill-page-version-projections.ts` -> `runPageVersionProjectionBackfill` -> `applyPageVersionProjectionBackfill`.                                                                          | Retain; release-gated historical Version canonicalization.       |
| `UpdateContentCache`                   | No request producer exists. The generic render worker has a page branch, but no production code publishes `PageRenderRequest`; source-locale page persistence already writes JSON/HTML/text synchronously. | Remove together with the dead page render queue/completion path. |
| `SaveDocument`                         | `editor-collab/handlers/page.ts:pageHandler.store` -> `savePageDocument`.                                                                                                                                  | Retain; exact loaded edit-hash CAS required.                     |
| `LoadDocument`                         | `pageHandler.load` and bounded page translation persistence -> `loadPageDocument`.                                                                                                                         | Retain; API integration audit pending.                           |
| `ListVersionProjectionBackfillTargets` | `scripts/ops/backfill-page-version-projections.ts` -> `runPageVersionProjectionBackfill` -> `listPageVersionProjectionBackfillTargets`.                                                                    | Retain; release-gated dry-run, apply cursor, and verify audit.   |
| `SaveTranslationDocument`              | `lib/page-translation-persistence.ts:savePreparedPageTranslationDocument` -> `savePageTranslationDocument`; invoked by locale persistence and rematerialization.                                           | Retain; API integration audit pending.                           |
| `RematerializeTranslationDocument`     | `lib/page-structure-reconcile.ts:prepareAndSavePageLocale` -> `rematerializePageTranslationDocument`; invoked only after a shared structure change or explicit operator repair.                            | Retain; bounded CAS, no background reconcile loop.               |
| `LoadTranslationDocument`              | Page translation persistence and structure reconcile flows -> `loadPageTranslationDocument`.                                                                                                               | Retain; API integration audit pending.                           |
| `ListTranslationLocales`               | Page translation persistence and structure reconcile flows -> `listPageTranslationLocales`.                                                                                                                | Retain; API integration audit pending.                           |
| `GetForRender`                         | No request producer exists. It is reachable only from the unproduced page render queue branch.                                                                                                             | Remove together with the dead page render queue/completion path. |

## `api.intra.v1.InternalPostService`

| RPC                       | Production call chain                                                                                                                      | Decision                                           |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------- |
| `UpdateContentCache`      | No runtime caller remains after the typed Block cutoff; source-document persistence owns any retained synchronous render-cache projection. | Remove; no generic render queue or worker remains. |
| `SaveDocument`            | Shared branch of `postHandler.store` -> `savePostDocument`.                                                                                | Retain; API integration audit pending.             |
| `LoadDocument`            | Shared branch of `postHandler.load` -> `loadPostDocument`.                                                                                 | Retain; API integration audit pending.             |
| `SaveTranslationDocument` | Locale branch of `postHandler.store` -> `savePostTranslationDocument`.                                                                     | Retain; API integration audit pending.             |
| `LoadTranslationDocument` | Locale branch of `postHandler.load` -> `loadPostTranslationDocument`.                                                                      | Retain; API integration audit pending.             |
| `ListTranslationLocales`  | No production caller or API-client function exists; only API tests call the implementation directly.                                       | Remove.                                            |
| `GetForRender`            | No runtime caller remains after the generic render worker removal.                                                                         | Remove.                                            |

## `api.intra.v1.InternalPrivacyService`

| RPC                       | Production call chain                                                                                                                      | Decision                                           |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------- |
| `UpdateContentCache`      | No runtime caller remains after the typed Block cutoff; legal source-document persistence retains its synchronous render-cache projection. | Remove; no generic render queue or worker remains. |
| `SaveDocument`            | Shared branch of `privacyHandler.store` -> `savePrivacyDocument`.                                                                          | Retain; API integration audit pending.             |
| `LoadDocument`            | Shared branch of `privacyHandler.load` -> `loadPrivacyDocument`.                                                                           | Retain; API integration audit pending.             |
| `SaveTranslationDocument` | Locale branch of `privacyHandler.store` -> `savePrivacyTranslationDocument`.                                                               | Retain; API integration audit pending.             |
| `LoadTranslationDocument` | Locale branch of `privacyHandler.load` -> `loadPrivacyTranslationDocument`.                                                                | Retain; API integration audit pending.             |
| `GetForRender`            | No runtime caller remains after the generic render worker removal.                                                                         | Remove.                                            |

## `api.intra.v1.InternalProgramEventService`

| RPC                       | Production call chain                                                                                                                      | Decision                                           |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------- |
| `UpdateContentCache`      | No runtime caller remains after the typed Block cutoff; source-document persistence owns any retained synchronous render-cache projection. | Remove; no generic render queue or worker remains. |
| `SaveTranslationDocument` | `programEventHandler.store` -> `saveProgramEventTranslationDocument`.                                                                      | Retain; API integration audit pending.             |
| `LoadTranslationDocument` | `programEventHandler.load` -> `loadProgramEventTranslationDocument`.                                                                       | Retain; API integration audit pending.             |
| `ListTranslationLocales`  | No production caller or API-client function exists; only API tests call the implementation directly.                                       | Remove.                                            |
| `GetForRender`            | No runtime caller remains after the generic render worker removal.                                                                         | Remove.                                            |

## `api.intra.v1.InternalReleaseService`

| RPC                       | Production call chain                                                                               | Decision                               |
| ------------------------- | --------------------------------------------------------------------------------------------------- | -------------------------------------- |
| `SaveDocument`            | Shared branch of `editor-collab/handlers/release.ts:releaseHandler.store` -> `saveReleaseDocument`. | Retain; API integration audit pending. |
| `LoadDocument`            | Shared branch of `releaseHandler.load` -> `loadReleaseDocument`.                                    | Retain; API integration audit pending. |
| `SaveTranslationDocument` | Locale branch of `releaseHandler.store` -> `saveReleaseTranslationDocument`.                        | Retain; API integration audit pending. |
| `LoadTranslationDocument` | Locale branch of `releaseHandler.load` -> `loadReleaseTranslationDocument`.                         | Retain; API integration audit pending. |

The former `GetReleaseIdForTrack` projection helper was removed. Release Track original-audio
projection no longer resolves or mutates a Release collaboration document; the API-owned queue
consumer must call `api.intra.v1.InternalFileIngestService.AttachTrackOriginalAudio` with the
verified File, ingest attempt, and exact current-File expectation.

## `api.intra.v1.InternalTermsService`

| RPC                       | Production call chain                                                                                                                      | Decision                                           |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------- |
| `UpdateContentCache`      | No runtime caller remains after the typed Block cutoff; legal source-document persistence retains its synchronous render-cache projection. | Remove; no generic render queue or worker remains. |
| `SaveDocument`            | Shared branch of `termsHandler.store` -> `saveTermsDocument`.                                                                              | Retain; API integration audit pending.             |
| `LoadDocument`            | Shared branch of `termsHandler.load` -> `loadTermsDocument`.                                                                               | Retain; API integration audit pending.             |
| `SaveTranslationDocument` | Locale branch of `termsHandler.store` -> `saveTermsTranslationDocument`.                                                                   | Retain; API integration audit pending.             |
| `LoadTranslationDocument` | Locale branch of `termsHandler.load` -> `loadTermsTranslationDocument`.                                                                    | Retain; API integration audit pending.             |
| `GetForRender`            | No runtime caller remains after the generic render worker removal.                                                                         | Remove.                                            |

## `api.intra.v1.InternalWorkService`

| RPC                       | Production call chain                                                                                                                      | Decision                                           |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------- |
| `UpdateContentCache`      | No runtime caller remains after the typed Block cutoff; source-document persistence owns any retained synchronous render-cache projection. | Remove; no generic render queue or worker remains. |
| `SaveDocument`            | Shared branch of `workHandler.store` -> `saveWorkDocument`.                                                                                | Retain; API integration audit pending.             |
| `LoadDocument`            | Shared branch of `workHandler.load` -> `loadWorkDocument`.                                                                                 | Retain; API integration audit pending.             |
| `SaveTranslationDocument` | Locale branch of `workHandler.store` -> `saveWorkTranslationDocument`.                                                                     | Retain; API integration audit pending.             |
| `LoadTranslationDocument` | Locale branch of `workHandler.load` -> `loadWorkTranslationDocument`.                                                                      | Retain; API integration audit pending.             |
| `ListTranslationLocales`  | No production caller or API-client function exists; only API tests call the implementation directly.                                       | Remove.                                            |
| `GetForRender`            | No runtime caller remains after the generic render worker removal.                                                                         | Remove.                                            |

## Open API

Oathkeeper permits `POST /api.open.v1.*` with optional authentication and an
anonymous fallback. The production caller for this layer is the web
application's generated Connect clients.

### `api.open.v1.ArtistService`

| RPC           | Production call chain                                                                               | Decision                           |
| ------------- | --------------------------------------------------------------------------------------------------- | ---------------------------------- |
| `List`        | `web/lib/actions/artist.ts:listArtistsForBlockAction` -> generated public artist client.            | Retain; integration audit pending. |
| `Get`         | `web/lib/queries/artist.ts:getArtistBySlug` and metadata queries -> generated public artist client. | Retain; integration audit pending. |
| `GetReleases` | Artist detail query -> generated public artist client.                                              | Retain; integration audit pending. |
| `GetWorks`    | Artist detail query -> generated public artist client.                                              | Retain; integration audit pending. |

### `api.open.v1.CategoryService`

| RPC    | Production call chain                                                            | Decision                           |
| ------ | -------------------------------------------------------------------------------- | ---------------------------------- |
| `List` | `web/lib/queries/taxonomy.ts:getCategories` -> generated public category client. | Retain; integration audit pending. |

### `api.open.v1.ClientService`

| RPC    | Production call chain                                                                        | Decision                           |
| ------ | -------------------------------------------------------------------------------------------- | ---------------------------------- |
| `Get`  | `web/lib/actions/client.ts:getClientsForBlockByIdsAction` -> generated public client client. | Retain; integration audit pending. |
| `List` | `web/lib/actions/client.ts:listClientsForBlockAction` -> generated public client client.     | Retain; integration audit pending. |

### `api.open.v1.FileService`

| RPC                 | Production call chain                                                                                                                                   | Decision                                         |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| `AuthorizeDownload` | `web/lib/queries/file-download-browser.ts` refreshes Content Block and Release Track downloads through the generated exact-relation public File client. | Retain; exact relation authority is intentional. |

### `api.open.v1.FormService`

| RPC              | Production call chain                                                                                                      | Decision                           |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| `CheckAccess`    | Form route metadata and `web/lib/actions/form.ts:checkFormAccessAction`. The response already carries the accessible form. | Retain; integration audit pending. |
| `Get`            | None. No public form client calls it; form retrieval is already part of `CheckAccess`.                                     | Remove.                            |
| `Submit`         | `web/lib/actions/form.ts:submitFormAction`.                                                                                | Retain; integration audit pending. |
| `VerifyPassword` | `web/lib/actions/form.ts:verifyFormPasswordAction`.                                                                        | Retain; integration audit pending. |
| `GetDashboard`   | `web/lib/actions/form.ts:getFormDashboardByShareAction`.                                                                   | Retain; integration audit pending. |

### `api.open.v1.FormatService`

| RPC    | Production call chain                                                              | Decision |
| ------ | ---------------------------------------------------------------------------------- | -------- |
| `List` | None. Web imports only the manage `FormatService`; no public format client exists. | Remove.  |
| `Get`  | None. Web imports only the manage `FormatService`; no public format client exists. | Remove.  |

### `api.open.v1.GenreService`

| RPC    | Production call chain                                                            | Decision |
| ------ | -------------------------------------------------------------------------------- | -------- |
| `List` | None. Web imports only the manage `GenreService`; no public genre client exists. | Remove.  |

### `api.open.v1.LabelService`

| RPC           | Production call chain                                                             | Decision                           |
| ------------- | --------------------------------------------------------------------------------- | ---------------------------------- |
| `List`        | Label block actions call the generated public label client.                       | Retain; integration audit pending. |
| `Get`         | Label detail, block, and metadata queries call the generated public label client. | Retain; integration audit pending. |
| `GetReleases` | None. The only public `getReleases` call belongs to `ArtistService`.              | Remove.                            |

### `api.open.v1.ManifestService`

| RPC   | Production call chain                                                               | Decision                           |
| ----- | ----------------------------------------------------------------------------------- | ---------------------------------- |
| `Get` | `web/lib/queries/metadata.ts:getSiteMetadataDocument` -> generated manifest client. | Retain; integration audit pending. |

### `api.open.v1.MapPlaceService`

| RPC        | Production call chain                                                                                    | Decision                           |
| ---------- | -------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| `Get`      | None. Public map blocks load places in batches through `GetByIds`; admin details use the manage service. | Remove.                            |
| `GetByIds` | `web/lib/actions/map-place.ts:getPublicMapPlacesByIdsAction`.                                            | Retain; integration audit pending. |
| `Search`   | None. Search/create flows use the manage browser client.                                                 | Remove.                            |

### `api.open.v1.MapThemeService`

| RPC            | Production call chain                                                                  | Decision                                      |
| -------------- | -------------------------------------------------------------------------------------- | --------------------------------------------- |
| `List`         | None. No generated public map-theme client call exists.                                | Remove.                                       |
| `ResolveByIds` | Public content hydration resolves each durable Theme reference, including deleted IDs. | Retain; ordered fallback integration pending. |
| `Resolve`      | `web/lib/actions/map-theme.ts:resolvePublicMapThemeAction`.                            | Retain; integration audit pending.            |

### `api.open.v1.MenuService`

| RPC   | Production call chain                                                                                   | Decision |
| ----- | ------------------------------------------------------------------------------------------------------- | -------- |
| `Get` | None. Web imports only the manage menu service; public navigation is supplied by `ManifestService.Get`. | Remove.  |

### `api.open.v1.PageService`

| RPC   | Production call chain                                                                       | Decision                           |
| ----- | ------------------------------------------------------------------------------------------- | ---------------------------------- |
| `Get` | Page route, homepage, manifest, and metadata queries call the generated public page client. | Retain; integration audit pending. |

### `api.open.v1.PostService`

| RPC               | Production call chain                                                                                           | Decision                           |
| ----------------- | --------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| `Get`             | Post route, metadata, and share-token queries call the generated public post client.                            | Retain; integration audit pending. |
| `List`            | Server and browser post listing queries plus member-profile post listing call the generated public post client. | Retain; integration audit pending. |
| `ListMapFeatures` | Server and browser map-block queries call the generated public post client.                                     | Retain; integration audit pending. |
| `Search`          | `web/lib/queries/post-browser.ts:searchPosts`.                                                                  | Retain; integration audit pending. |

### `api.open.v1.PrivacyService`

| RPC    | Production call chain                                                        | Decision                           |
| ------ | ---------------------------------------------------------------------------- | ---------------------------------- |
| `Get`  | Privacy routes and browser queries call the generated public privacy client. | Retain; integration audit pending. |
| `List` | `web/lib/queries/privacy-browser.ts:listPrivacyVersions`.                    | Retain; integration audit pending. |

### `api.open.v1.ProgramEventService`

| RPC    | Production call chain                                                            | Decision                           |
| ------ | -------------------------------------------------------------------------------- | ---------------------------------- |
| `Get`  | Program-event route and metadata queries call the generated public event client. | Retain; integration audit pending. |
| `List` | Program-event server and browser queries call the generated public event client. | Retain; integration audit pending. |

### `api.open.v1.ProgramEventSeriesService`

| RPC    | Production call chain                                                          | Decision                           |
| ------ | ------------------------------------------------------------------------------ | ---------------------------------- |
| `Get`  | Program-event series detail query calls the generated public series client.    | Retain; integration audit pending. |
| `List` | Program-event browser taxonomy query calls the generated public series client. | Retain; integration audit pending. |

### `api.open.v1.ProgramEventTypeService`

| RPC    | Production call chain                                                        | Decision                           |
| ------ | ---------------------------------------------------------------------------- | ---------------------------------- |
| `List` | Program-event browser taxonomy query calls the generated public type client. | Retain; integration audit pending. |

### `api.open.v1.ReleaseService`

| RPC    | Production call chain                                                            | Decision                           |
| ------ | -------------------------------------------------------------------------------- | ---------------------------------- |
| `Get`  | Release detail and metadata queries call the generated public release client.    | Retain; integration audit pending. |
| `List` | Release server/browser listing queries call the generated public release client. | Retain; integration audit pending. |

### `api.open.v1.SeriesService`

| RPC         | Production call chain                                                                          | Decision                                |
| ----------- | ---------------------------------------------------------------------------------------------- | --------------------------------------- |
| `Get`       | `web/lib/queries/series.ts:getPublicSeries` serves `/series/[idOrSlug]`.                       | Retain; consumer and provider verified. |
| `List`      | `web/lib/queries/series.ts:listPublicSeriesOptions`.                                           | Retain; consumer and provider verified. |
| `ListPosts` | None. Ordered Series Posts use `PostService.List` with `series_id` and `series_order` instead. | Remove.                                 |

### `api.open.v1.ShareLinkService`

| RPC        | Production call chain                                             | Decision                           |
| ---------- | ----------------------------------------------------------------- | ---------------------------------- |
| `Validate` | `web/app/s/[token]/route.ts` validates and redirects share links. | Retain; integration audit pending. |

### `api.open.v1.SitemapService`

| RPC           | Production call chain                            | Decision                           |
| ------------- | ------------------------------------------------ | ---------------------------------- |
| `GetDocument` | `web/lib/queries/sitemap.ts:getSitemapDocument`. | Retain; integration audit pending. |

### `api.open.v1.StyleService`

| RPC    | Production call chain                                                            | Decision |
| ------ | -------------------------------------------------------------------------------- | -------- |
| `List` | None. Web imports only the manage `StyleService`; no public style client exists. | Remove.  |

### `api.open.v1.NewsletterService`

| RPC           | Target consumer (not yet implemented)                                            | Decision                                          |
| ------------- | -------------------------------------------------------------------------------- | ------------------------------------------------- |
| `Unsubscribe` | `/unsubscribe?token=...` will call the generated public newsletter token client. | Retain; API/Web implementation and audit pending. |

### `api.open.v1.TagService`

| RPC    | Production call chain                  | Decision                           |
| ------ | -------------------------------------- | ---------------------------------- |
| `List` | `web/lib/queries/taxonomy.ts:getTags`. | Retain; integration audit pending. |

### `api.open.v1.TermsService`

| RPC    | Production call chain                                                    | Decision                           |
| ------ | ------------------------------------------------------------------------ | ---------------------------------- |
| `Get`  | Terms routes and browser queries call the generated public terms client. | Retain; integration audit pending. |
| `List` | `web/lib/queries/terms-browser.ts:listTermsVersions`.                    | Retain; integration audit pending. |

### `api.open.v1.TrackService`

| RPC   | Production call chain                                                            | Decision |
| ----- | -------------------------------------------------------------------------------- | -------- |
| `Get` | None. Web imports only the manage `TrackService`; no public track client exists. | Remove.  |

### `api.open.v1.MemberService`

The provider is the `MemberService` type in
`api/internal/service/public/user.go`. The filename is a legacy implementation
path; it does not restore a `UserService` contract or mix account authority into
the public member surface.

| RPC               | Production call chain                                                                                                                                                                      | Decision                           |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------- |
| `GetPublicMember` | `web/lib/queries/user.ts:getUserProfileView` and `web/lib/queries/metadata.ts:getMemberMetadata` call `createPublicMemberClient`; the response contains public Member profile fields only. | Retain; integration audit pending. |
| `ListAuthors`     | `web/lib/queries/user.ts:listAuthors` calls `createPublicMemberClient`; author identity is a `MemberSummary` and the provider counts durable post authorship.                              | Retain; integration audit pending. |

### `api.open.v1.AccountService`

The provider is the distinct `AccountService` type in
`api/internal/service/public/user.go`. These token/email-authenticated lifecycle
methods do not expose a public profile and do not accept a Member identifier as
proof of account ownership.

| RPC                      | Production call chain                                                                            | Decision                           |
| ------------------------ | ------------------------------------------------------------------------------------------------ | ---------------------------------- |
| `ConfirmAccountDeletion` | `web/lib/actions/account.ts:confirmAccountDeletionAction` calls `createPublicAccountClient`.     | Retain; integration audit pending. |
| `CancelAccountDeletion`  | `web/lib/actions/account.ts:cancelAccountDeletionAction` calls `createPublicAccountClient`.      | Retain; integration audit pending. |
| `RequestAccountRecovery` | `web/lib/actions/account.ts:requestAccountRecoveryFormAction` calls `createPublicAccountClient`. | Retain; integration audit pending. |
| `ConfirmAccountRecovery` | `web/lib/actions/account.ts:confirmAccountRecoveryAction` calls `createPublicAccountClient`.     | Retain; integration audit pending. |

The former `api.open.v1.UserService` declaration, `user.proto`, and generated
client/server surface are intentionally removed with no compatibility alias.
That service-symbol removal is retained here as hard-cut history. Its individual
operations are not recorded as removed behavior: profile/author operations were
renamed under `MemberService`, while deletion/recovery operations moved under
`AccountService` and have direct production consumers.

### `api.open.v1.WorkService`

| RPC               | Production call chain                                                                        | Decision                           |
| ----------------- | -------------------------------------------------------------------------------------------- | ---------------------------------- |
| `Get`             | Work detail, metadata, and share-token queries call the generated public work client.        | Retain; integration audit pending. |
| `List`            | Work server/browser listing queries and block actions call the generated public work client. | Retain; integration audit pending. |
| `ListMapFeatures` | Work map-block server/browser queries call the generated public work client.                 | Retain; integration audit pending. |

## Manage API

The machine-readable procedure list and per-RPC review/test state live in
`docs/rpc-inventory.json`. The table below records the implementation and real
production entry points inspected for each service. Oathkeeper policy is derived
from each method's protobuf access option; manage procedures are never inferred
from method names.

| Service                                   | API implementation evidence                                                        | Direct production caller evidence                                                                                      | Usage decision                                                                                           |
| ----------------------------------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| `api.manage.v1.AdminService`              | `api/internal/service/admin.go`                                                    | `web/lib/queries/admin-browser.ts`, post/page/series/work/site-setting actions                                         | All declared RPCs retained.                                                                              |
| `api.manage.v1.AIService`                 | `api/internal/service/ai.go`                                                       | `web/lib/ai/chat.ts`, `web/lib/ai/metadata-jobs.ts`                                                                    | All declared RPCs retained.                                                                              |
| `api.manage.v1.ArtistService`             | `api/internal/service/artist.go`                                                   | `web/lib/actions/artist.ts`, `web/lib/queries/artist-browser.ts`                                                       | All declared RPCs retained.                                                                              |
| `api.manage.v1.AudienceService`           | `api/internal/service/audience.go`                                                 | `web/lib/actions/audience.ts`                                                                                          | Archive/restore replace the removed delete RPC.                                                          |
| `api.manage.v1.CampaignService`           | `api/internal/service/campaign.go`                                                 | `web/lib/actions/campaign.ts`                                                                                          | All declared RPCs retained.                                                                              |
| `api.manage.v1.CategoryService`           | `api/internal/service/category.go`                                                 | `web/lib/actions/category.ts`                                                                                          | All declared RPCs retained.                                                                              |
| `api.manage.v1.ClientService`             | `api/internal/service/client.go`                                                   | `web/lib/actions/client.ts`, `web/lib/queries/client-browser.ts`                                                       | All declared RPCs retained.                                                                              |
| `api.manage.v1.CommentService`            | `api/internal/service/comment.go`                                                  | `web/lib/actions/comment.ts`                                                                                           | All declared RPCs retained.                                                                              |
| `api.manage.v1.EmailLayoutService`        | `api/internal/service/email_layout.go`                                             | `web/lib/actions/email-layout.ts`, email-layout admin queries                                                          | All declared RPCs retained.                                                                              |
| `api.manage.v1.EmailSuppressionService`   | `api/internal/service/email_suppression.go`                                        | `web/lib/actions/email-suppression.ts`                                                                                 | All declared RPCs retained.                                                                              |
| `api.manage.v1.EmailTemplateService`      | `api/internal/service/email_template.go`                                           | `web/lib/actions/email-template.ts`, email-template admin queries                                                      | All declared RPCs retained.                                                                              |
| `api.manage.v1.FileService`               | `api/internal/service/file.go`                                                     | `web/lib/actions/file.ts`, program-event media hydration, site settings, immersive-scene optimization                  | All declared RPCs retained.                                                                              |
| `api.manage.v1.FormService`               | `api/internal/service/form.go`                                                     | `web/lib/actions/form.ts`, form server/browser queries                                                                 | All declared RPCs retained.                                                                              |
| `api.manage.v1.FormatService`             | `api/internal/service/format.go`                                                   | `web/lib/actions/format.ts`                                                                                            | All declared RPCs retained.                                                                              |
| `api.manage.v1.GenreService`              | `api/internal/service/genre.go`                                                    | `web/lib/actions/genre.ts`                                                                                             | All declared RPCs retained.                                                                              |
| `api.manage.v1.LabelService`              | `api/internal/service/label.go`                                                    | `web/lib/actions/label.ts`, label server/browser queries                                                               | All declared RPCs retained.                                                                              |
| `api.manage.v1.MailAdapterService`        | `api/internal/service/mail_adapter.go`                                             | `web/lib/actions/mail-adapter.ts`                                                                                      | Current RPCs retained; unused `Get` removed.                                                             |
| `api.manage.v1.MapPlaceService`           | `api/internal/service/map_place.go`                                                | `web/lib/actions/map-place.ts`, `web/lib/api/map-place-browser-client.ts`, location selectors                          | All declared RPCs retained.                                                                              |
| `api.manage.v1.MapThemeService`           | `api/internal/service/map_theme.go`                                                | `web/lib/actions/map-theme.ts`                                                                                         | Dedicated default selection retained; generic update removed because collaborative CAS owns theme edits. |
| `api.manage.v1.MenuService`               | `api/internal/service/menu.go`                                                     | menu actions and server/browser queries                                                                                | All declared RPCs retained.                                                                              |
| `api.manage.v1.PageService`               | `api/internal/service/page.go`                                                     | page actions/queries, version-history API routes and drawer                                                            | All declared RPCs retained.                                                                              |
| `api.manage.v1.PostService`               | `api/internal/service/post.go`                                                     | `web/lib/actions/post.ts`, post queries, version-history API routes and drawer                                         | All declared RPCs retained.                                                                              |
| `api.manage.v1.PrivacyService`            | `api/internal/service/privacy.go`                                                  | `web/lib/actions/privacy.ts`, `web/lib/queries/privacy.ts`                                                             | All declared RPCs retained.                                                                              |
| `api.manage.v1.ProgramEventService`       | `api/internal/service/program_event.go`                                            | `web/lib/actions/program-event.ts`, program-event admin queries                                                        | All declared RPCs retained.                                                                              |
| `api.manage.v1.ProgramEventSeriesService` | `api/internal/service/program_event.go`                                            | program-event series actions and queries                                                                               | All declared RPCs retained.                                                                              |
| `api.manage.v1.ProgramEventTypeService`   | `api/internal/service/program_event.go`                                            | program-event type actions and queries                                                                                 | All declared RPCs retained.                                                                              |
| `api.manage.v1.ReleaseService`            | `api/internal/service/release.go`                                                  | `web/lib/actions/release.ts`, release server/browser queries                                                           | All declared RPCs retained.                                                                              |
| `api.manage.v1.SeriesService`             | `api/internal/service/series.go`                                                   | `web/lib/actions/series.ts`, series server/browser queries                                                             | All declared RPCs retained; exact consumers and concentrated provider coverage verified.                 |
| `api.manage.v1.ShareLinkService`          | `api/internal/service/share_link.go`                                               | `web/lib/actions/share-link.ts`                                                                                        | All declared RPCs retained.                                                                              |
| `api.manage.v1.SiteSettingService`        | `api/internal/service/site_setting.go`                                             | `web/lib/actions/site-setting.ts`, site-setting queries                                                                | All declared RPCs retained.                                                                              |
| `api.manage.v1.StyleService`              | `api/internal/service/style.go`                                                    | `web/lib/actions/style.ts`                                                                                             | Current RPCs retained; unused `GetStyle` removed.                                                        |
| `api.manage.v1.TagService`                | `api/internal/service/tag.go`                                                      | `web/lib/actions/tag.ts`                                                                                               | All declared RPCs retained.                                                                              |
| `api.manage.v1.TermsService`              | `api/internal/service/terms.go`                                                    | `web/lib/actions/terms.ts`, `web/lib/queries/terms.ts`                                                                 | All declared RPCs retained.                                                                              |
| `api.manage.v1.TrackService`              | `api/internal/service/track.go`                                                    | `web/lib/actions/track.ts`                                                                                             | All declared RPCs retained.                                                                              |
| `api.manage.v1.TranslationService`        | `api/internal/service/translation_service.go`                                      | translation admin pages and translation editor panels/hooks                                                            | All declared RPCs retained.                                                                              |
| `api.manage.v1.MemberService`             | `api/internal/service/member_service.go` (`MemberService` Member/profile provider) | `web/lib/auth.ts`; member/profile, locale, cookie-consent, tag actions/queries; authenticated/admin Member UI consumer | Member/profile owner; lean session, self-detail, settings and admin reads have distinct projections.     |
| `api.manage.v1.AccountService`            | `api/internal/service/account_service.go`                                          | account, newsletter, email, identity, session, and account-admin Web paths                                             | Kratos/account and Identity-owned newsletter mutation owner; per-RPC consumer audit pending.             |
| `api.manage.v1.WorkService`               | `api/internal/service/work.go`                                                     | `web/lib/actions/work.ts`, work queries, version-history API routes and drawer                                         | All declared RPCs retained.                                                                              |

The former `api.manage.v1.UserService`, `proto/api/manage/v1/user.proto`, and
generated service surface are intentionally removed with no compatibility
alias. The old `api/internal/service/user.go` provider reference is therefore
deleted from current evidence: Member operations are implemented by
`member_service.go`, and Kratos/account operations by `account_service.go`.
`GetCurrentSession` embeds only the bounded account fields needed by global
login and role handling. `GetMember` and `ListMembersAdmin` may embed wider
`AccountSummary` or `AccountAdminDetails` projections to avoid consumer N+1
calls; only `AccountService` mutates account/identity state.

The manage `MemberService` owns profile, Member preferences, avatar, tags, and
the session/self-detail/settings/admin projections. Its direct consumers include
`web/lib/auth.ts`, the member/profile and tag-assignment branches of
`web/lib/actions/user.ts`, member queries, and the locale/cookie-consent/tag
actions. `AccountService` owns
Identity-scoped newsletter mutation, Kratos-backed email, provider,
credential/session security, role/ban state, registration provisioning, and
deletion lifecycle; its direct consumers are the newsletter/account/email/
identity/session actions and the account-only branches of
`web/lib/actions/user.ts`. The providers implement all declared methods. No
`MemberService.SetMemberTags` is consumed by the Admin Member profile update
action and its consumer test covers both assignment and clearing. No retained
`AccountService.GetAccount` procedure exists in the current contract.

For subtraction history, retain the removal of the old service symbol and the
standalone procedures whose result was folded into a new aggregate (for example,
old My-section and cookie-consent reads). Do not retain obsolete per-RPC
"Remove" decisions when the same operation survives under `MemberService` or
`AccountService`; the current procedure-level evidence and retention authority
remain in `docs/rpc-inventory.json`.
