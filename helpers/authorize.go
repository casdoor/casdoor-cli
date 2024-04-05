package helpers

// HasRequiredGroup checks if the user has at least one of the required permissions
// in their claims. Returns true if the user has any of the target permissions;
// otherwise, returns false.
func HasRequiredGroup(claimsPermission []string, targetPermission []string) bool {
	for _, claimsRole := range claimsPermission {
		for _, targetRole := range targetPermission {
			if claimsRole == targetRole {
				return true
			}
		}
	}
	return false
}
