# Research: Reduce Presigned URL Expiration to 60 Seconds

**Feature**: `011-reduce-presigned-url-expiration`  
**Date**: 2026-10-23  

---

## 1. Expiration Duration Configuration for Direct Upload URLs

### Decision
Change the constant `expirationSeconds` in `internal/library/service/file_service.go` (`GenerateUploadURL`) from `900` (15 minutes) to `60` (60 seconds). Return `"expires_in": 60` in `PresignedUploadResponse`.

### Rationale
- A 60-second window is ideal for direct web uploads: the frontend requests the URL immediately when the user clicks "Enviar" and starts the HTTP `PUT` transfer in milliseconds.
- Closing the window from 15 minutes to 60 seconds minimizes the attack surface if an upload URL is accidentally leaked or intercepted before ingestion.

### Alternatives Considered
- *30 seconds*: Too aggressive; mobile or congested networks might experience DNS lookup or TCP handshake delays exceeding 30s before the first byte arrives.
- *300 seconds (5 minutes)*: Excessively long for automated direct client-side uploads where the client is already holding the file in memory or file input.

---

## 2. Expiration Duration for Download and Preview URLs

### Decision
Update the expiration duration in `internal/library/service/file_service.go` (`GetFilesByCreator` and `GetFilesByCreatorPaginated`) from `5 * time.Minute` (300 seconds) to `60 * time.Second` (60 seconds).

### Rationale
- Comports directly with Constitution Principle I (Content Protection).
- Files in the creator's library are sensitive private assets. When a user navigates to the library or clicks preview/download, the client requests a fresh list or uses the current presigned URL immediately.
- A 60-second expiration ensures that even if an authorized user extracts the direct S3 URL from browser network inspect tools, that URL becomes dead within 1 minute, preventing unauthorized redistribution.

### Alternatives Considered
- *Leaving download links at 5 minutes*: Inconsistent with the requirement to enforce a strict 60-second window across presigned access.
- *Permanent public URLs*: Directly violates Constitution Principle I.

---

## 3. Storage Layer and AWS S3 SDK Behavior

### Decision
Verify and maintain the pass-through of `expirationSeconds` in `pkg/storage/s3.go`:
- `GeneratePresignedPutURL(ctx context.Context, key, contentType string, expirationSeconds int) (string, error)`
- `GeneratePreviewLinkWithExpiration(key, contentType string, expirationSeconds int) string`
- `GenerateDownloadLinkWithExpiration(key string, expirationSeconds int) string`

Ensure that the S3 presigner client configures `opts.Expires = time.Duration(expirationSeconds) * time.Second`.

### AWS S3 Protocol Behavior for Long Transfers
- **Evaluation at Handshake**: AWS S3 validates the expiration timestamp (`X-Amz-Expires` and `X-Amz-Date`) at the time the HTTP request headers are received.
- If a client with a 10MB file starts the upload at second 10, the transfer continues to completion even if streaming takes 90 seconds, as long as the TCP connection was opened and the HTTP `PUT` request was accepted before the 60 seconds elapsed.

---

## 4. Frontend Resilience and Retry UX

### Decision
Enhance error handling in `frontend/src/lib/s3-upload.ts` and `frontend/src/pages/files/FileLibraryPage.tsx`:
- When an upload fails with HTTP status 403 or network failure, explicitly detect possible URL expiration.
- If an upload expires or fails, retain the user's form inputs (name, description, selected file) so they can click "Tentar Novamente", instantly generating a fresh 60-second URL and executing the upload without re-selecting the file.
