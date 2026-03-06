<<<<<<< SEARCH
			// Don't log requests to the audit endpoint itself to avoid loops
			if path == "/activitylist" && method == "GET" {
				return
			}
=======
			// Don't log requests to the audit endpoint itself to avoid loops
			if path == "/activitylist" {
				return
			}
>>>>>>> REPLACE
