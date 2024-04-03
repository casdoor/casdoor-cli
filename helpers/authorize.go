package helpers

func HasRequiredRole(claimsRoles []string, targetRoles []string) bool {
	for _, claimsRole := range claimsRoles {
		for _, targetRole := range targetRoles {
			if claimsRole == targetRole {
				return true
			}
		}
	}
	return false
}
