# Requirements: Illoominate

**Defined:** 2026-01-21
**Core Value:** Native-feeling feedback submission with transparent, startup-friendly pricing

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Landing Page

- [ ] **LAND-01**: Landing page explains Illoominate value proposition
- [ ] **LAND-02**: User can sign up for waitlist with email
- [ ] **LAND-03**: User receives confirmation email after waitlist signup
- [ ] **LAND-04**: Admin can view and export waitlist data

### Authentication

- [ ] **AUTH-01**: User can sign up with email and password
- [ ] **AUTH-02**: User can sign in with Google OAuth
- [ ] **AUTH-03**: User can sign in with GitHub OAuth
- [ ] **AUTH-04**: User can reset password via email link
- [ ] **AUTH-05**: User session persists across browser refresh

### Workspace & Apps

- [ ] **WORK-01**: User can create a workspace (organization)
- [ ] **WORK-02**: Workspace can contain multiple apps/products
- [ ] **WORK-03**: Each app has unique API keys for SDK authentication
- [ ] **WORK-04**: Workspace owner can add and remove team members

### Feedback Submission

- [ ] **SUBM-01**: User can submit feedback with type (Bug, Feature Request, UI/UX)
- [ ] **SUBM-02**: Feedback includes title and description
- [ ] **SUBM-03**: User can attach images/screenshots to feedback
- [ ] **SUBM-04**: SDK automatically captures device and app metadata

### Feedback Organization

- [ ] **ORG-01**: Admin can create boards to organize feedback
- [ ] **ORG-02**: Feedback has status workflow (Open → Under Review → Planned → In Progress → Complete)
- [ ] **ORG-03**: Admin can filter feedback by status, type, votes, date
- [ ] **ORG-04**: Admin can search feedback by text
- [ ] **ORG-05**: Admin can add tags/labels to feedback

### Voting & Engagement

- [ ] **VOTE-01**: User can upvote feedback items
- [ ] **VOTE-02**: Vote counts are visible on feedback items
- [ ] **VOTE-03**: User can comment on feedback items
- [ ] **VOTE-04**: User can subscribe to feedback items for updates

### Public Roadmap

- [ ] **ROAD-01**: Users can view public roadmap showing planned/in-progress/shipped items
- [ ] **ROAD-02**: Admin has private roadmap layer for internal planning
- [ ] **ROAD-03**: Admin can publish changelog when features ship
- [ ] **ROAD-04**: Roadmap can be embedded on external sites

### Notifications

- [ ] **NOTF-01**: User receives email when features they voted for ship
- [ ] **NOTF-02**: User has in-app notification center
- [ ] **NOTF-03**: User can configure notification preferences
- [ ] **NOTF-04**: User receives email when someone replies to their comment

### iOS SDK

- [ ] **IOS-01**: iOS SDK provides native SwiftUI UI for submitting feedback
- [ ] **IOS-02**: iOS SDK passes user identity from host app
- [ ] **IOS-03**: iOS SDK queues feedback offline and syncs when connected
- [ ] **IOS-04**: iOS SDK captures device metadata automatically

### Web SDK

- [ ] **WEB-01**: Web SDK provides JavaScript API for submitting feedback
- [ ] **WEB-02**: Web SDK passes user identity from host app
- [ ] **WEB-03**: Web SDK provides optional embeddable widget
- [ ] **WEB-04**: Web SDK captures browser metadata automatically

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Additional SDKs

- **ASDK-01**: Android SDK (Kotlin) with native UI
- **ASDK-02**: React Native SDK for cross-platform
- **ASDK-03**: Flutter SDK for cross-platform
- **ASDK-04**: Browse and vote within native SDKs

### Integrations

- **INTG-01**: Jira integration for syncing feedback to issues
- **INTG-02**: Linear integration for syncing feedback to issues
- **INTG-03**: GitHub integration for syncing feedback to issues
- **INTG-04**: Slack integration for notifications
- **INTG-05**: Discord integration for notifications
- **INTG-06**: Webhooks for custom automation

### AI Features

- **AI-01**: AI duplicate detection on submission
- **AI-02**: AI auto-categorization suggestions
- **AI-03**: AI-powered search and similarity

### Advanced Features

- **ADV-01**: Import from Canny/UserVoice/Fider
- **ADV-02**: SSO/SAML authentication
- **ADV-03**: Advanced analytics dashboard
- **ADV-04**: Custom fields on feedback
- **ADV-05**: Bulk actions (merge, close, tag)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Real-time chat/support | Different product category, not feedback tool |
| Video attachments | Storage/bandwidth costs, complexity |
| Mobile app for admins | Web dashboard sufficient for v1 |
| Forums/discussions | Scope creep into community platforms |
| Full project management | Sprints, assignments, time tracking is Linear/Jira territory |
| Gamification/points | Badges/leaderboards attract wrong behavior |
| Enterprise SSO/SCIM | Months of work for non-target market |
| Complex voting systems | Weighted votes, limits, resets overcomplicate |
| Anonymous feedback | Identity required for user linking |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| LAND-01 | Phase 1 | Pending |
| LAND-02 | Phase 1 | Pending |
| LAND-03 | Phase 1 | Pending |
| LAND-04 | Phase 1 | Pending |
| AUTH-01 | Phase 3 | Pending |
| AUTH-02 | Phase 3 | Pending |
| AUTH-03 | Phase 3 | Pending |
| AUTH-04 | Phase 3 | Pending |
| AUTH-05 | Phase 3 | Pending |
| WORK-01 | Phase 4 | Pending |
| WORK-02 | Phase 4 | Pending |
| WORK-03 | Phase 4 | Pending |
| WORK-04 | Phase 4 | Pending |
| SUBM-01 | Phase 7 | Pending |
| SUBM-02 | Phase 7 | Pending |
| SUBM-03 | Phase 7 | Pending |
| SUBM-04 | Phase 7 | Pending |
| ORG-01 | Phase 8 | Pending |
| ORG-02 | Phase 8 | Pending |
| ORG-03 | Phase 8 | Pending |
| ORG-04 | Phase 8 | Pending |
| ORG-05 | Phase 8 | Pending |
| VOTE-01 | Phase 9 | Pending |
| VOTE-02 | Phase 9 | Pending |
| VOTE-03 | Phase 9 | Pending |
| VOTE-04 | Phase 9 | Pending |
| ROAD-01 | Phase 11 | Pending |
| ROAD-02 | Phase 11 | Pending |
| ROAD-03 | Phase 11 | Pending |
| ROAD-04 | Phase 11 | Pending |
| NOTF-01 | Phase 12 | Pending |
| NOTF-02 | Phase 12 | Pending |
| NOTF-03 | Phase 12 | Pending |
| NOTF-04 | Phase 12 | Pending |
| IOS-01 | Phase 10 | Pending |
| IOS-02 | Phase 10 | Pending |
| IOS-03 | Phase 10 | Pending |
| IOS-04 | Phase 10 | Pending |
| WEB-01 | Phase 13 | Pending |
| WEB-02 | Phase 13 | Pending |
| WEB-03 | Phase 13 | Pending |
| WEB-04 | Phase 13 | Pending |

**Coverage:**
- v1 requirements: 40 total
- Mapped to phases: 40
- Unmapped: 0

**Note:** Phases 2, 5, and 6 are infrastructure phases that enable requirements but don't directly implement any. They provide the foundation (database, API, dashboard) that requirement-implementing phases build upon.

---
*Requirements defined: 2026-01-21*
*Last updated: 2026-01-21 after roadmap creation (13 phases)*
