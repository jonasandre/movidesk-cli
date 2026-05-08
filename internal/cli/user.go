package cli

// EnvUser short-circuits tenant.DefaultUser when set.
const EnvUser = "MOVIDESK_USER"

// injectCreatedBy adds {"createdBy":{"id": userID}} to body when:
//   - userID != ""
//   - body has no "createdBy" key
//
// An explicit createdBy from --set/--file always wins.
func injectCreatedBy(body map[string]any, userID string) {
	if userID == "" || body == nil {
		return
	}
	if _, ok := body["createdBy"]; ok {
		return
	}
	body["createdBy"] = map[string]any{"id": userID}
}
