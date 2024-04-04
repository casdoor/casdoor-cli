package helpers

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
