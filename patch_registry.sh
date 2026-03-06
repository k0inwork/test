<<<<<<< SEARCH
	r.GET("/admin/services", func(c *gin.Context) {
		mu.RLock()
		defer mu.RUnlock()
=======
	r.GET("/capabilities/:name", func(c *gin.Context) {
		name := c.Param("name")
		mu.RLock()
		defer mu.RUnlock()
		now := time.Now()
		for _, s := range registry {
			if now.Sub(s.LastUpdate) < 60*time.Second && s.Enabled {
				for _, cap := range s.Capabilities {
					if cap.Name == name {
						if len(cap.Endpoints) > 0 {
							// Return the service base endpoint + the capability endpoint
							c.JSON(http.StatusOK, gin.H{"endpoint": s.Endpoint + cap.Endpoints[0]})
							return
						}
					}
				}
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "capability not found"})
	})

	r.GET("/admin/services", func(c *gin.Context) {
		mu.RLock()
		defer mu.RUnlock()
>>>>>>> REPLACE
