# Feature Specification: Reduce Presigned URL Expiration to 60 Seconds

**Feature Branch**: `011-reduce-presigned-url-expiration`  
**Created**: 2026-10-23  
**Status**: Draft  
**Input**: User description: "Reduza o tempo de expiracao da URL pré assinada para 60 segundos."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Secure Short-Lived Direct Upload Window (Priority: P1)

As a creator uploading digital content, I want the temporary authorization link generated for uploading my file to remain active for only 60 seconds, so that the transfer window is strictly limited and unauthorized third parties cannot reuse the authorization if intercepted.

**Why this priority**: Directly enforces the project's content protection principle by closing unnecessary authorization windows for direct file ingestion.

**Independent Test**: Request an upload authorization, initiate the file upload within 60 seconds (succeeds), and verify that attempting an upload with the same link after 60 seconds is immediately rejected.

**Acceptance Scenarios**:

1. **Given** an authenticated creator initiating a file upload, **When** the upload authorization is requested, **Then** the generated authorization link is configured with a validity period of 60 seconds and the returned metadata communicates an expiration duration of 60 seconds.
2. **Given** a valid upload authorization generated less than 60 seconds ago, **When** the file transfer is started, **Then** the storage service accepts the file data.
3. **Given** an upload authorization generated more than 60 seconds ago, **When** a transfer attempt is made, **Then** the storage service rejects the transfer as expired.

---

### User Story 2 - Secure Short-Lived File Download and Preview Access (Priority: P2)

As a content owner or authorized buyer, when I access or preview a digital file in the platform, I want the temporary link granting access to the file to expire after 60 seconds, so that links shared or extracted from the browser cannot be redistributed or accessed publicly.

**Why this priority**: Protects creator assets from unauthorized distribution by ensuring links to private files are ephemeral and strictly single-use/short-lived.

**Independent Test**: Request a download or preview link for an owned file, verify the file opens/downloads immediately, and confirm that reusing the exact link 61 seconds later fails with access denied.

**Acceptance Scenarios**:

1. **Given** an authenticated user requesting file details or a download link, **When** the access link is generated, **Then** the link validity is constrained to exactly 60 seconds.
2. **Given** an access link generated within the last 60 seconds, **When** opened in a browser or download client, **Then** the file content is served.
3. **Given** an access link generated more than 60 seconds ago, **When** opened, **Then** the request is rejected with an expired/unauthorized response.

---

### User Story 3 - Expired Transfer Feedback and Immediate Recovery (Priority: P3)

As a creator or buyer whose network delay or hesitation causes an authorization link to expire before the transfer starts, I want the application to clearly inform me that the link expired and allow me to immediately request a fresh link without losing my context.

**Why this priority**: Balances strict 60-second security with smooth user experience so users are not blocked if a link expires.

**Independent Test**: Simulate an expired upload or download attempt, verify clear feedback indicates link expiration, and confirm that requesting a retry issues a new 60-second valid link immediately.

**Acceptance Scenarios**:

1. **Given** an upload or download attempt that fails due to link expiration, **When** the failure occurs, **Then** the user receives clear feedback stating the link has expired.
2. **Given** an expired transfer prompt, **When** the user triggers a retry, **Then** a fresh 60-second authorization is generated without requiring the user to re-authenticate or re-enter file metadata.

---

### Edge Cases

- **Large file uploads spanning over 60 seconds**: In cloud object storage protocols, the expiration timestamp is validated at the moment the transfer request begins (handshake/HTTP connection establishment). Transfers initiated within the 60-second window must continue until completion, provided the connection was opened before second 60.
- **Client clock desynchronization**: If the user's device clock is skewed, validation is strictly governed by the server and cloud storage provider's coordinated time; the UI must base any countdowns or refresh prompts on the duration returned by the server (`60` seconds) rather than device local clock timestamps.
- **Immediate retry spamming**: Repeated rapid requests for fresh URLs must remain subject to existing user rate limits while still allowing legitimate retries.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST configure temporary direct upload authorization links with a time-to-live of exactly 60 seconds.
- **FR-002**: The upload authorization response payload MUST report an expiration duration of 60 seconds to client applications.
- **FR-003**: The system MUST configure temporary download and preview links for files in the library with a time-to-live of exactly 60 seconds.
- **FR-004**: The storage access layer MUST reject any upload or retrieval request initiated after the 60-second expiration threshold has elapsed.
- **FR-005**: The client interface MUST clearly inform the user if a file transfer failed due to authorization expiration and provide a direct action to obtain a fresh link.

### Key Entities *(include if feature involves data)*

- **Temporary Authorization Request**: Data representing the intent to upload or download a file, containing file identification, intended MIME type, creator identity, and a designated 60-second expiration duration.
- **Temporary Authorization Response**: Communication payload returning the signed access link, required headers, unique resource identifier, and remaining validity in seconds (60 seconds).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of generated upload and download authorization links reject access attempts initiated after 60 seconds from issuance.
- **SC-002**: Legitimate transfers initiated within 60 seconds of issuance achieve 100% successful connection establishment under standard network conditions.
- **SC-003**: In the event of an expired authorization, users can obtain a replacement authorization in less than 2 seconds with a single retry action.
- **SC-004**: Zero perpetual or long-lived direct links (>60 seconds) exist for private creator library assets.

## Assumptions

- Cloud storage provider evaluates URL expiration at request initiation time, meaning that transfers started before the 60th second will not be terminated mid-stream once the payload begins transmitting.
- Standard user internet connection is adequate to initiate the HTTP handshake for an upload or download within 60 seconds of clicking the action.
- The 60-second limit applies to private library file uploads, previews, and downloads.
