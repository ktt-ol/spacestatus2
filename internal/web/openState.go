package web

import (
	"github.com/ktt-ol/status2/internal/state"
	"github.com/gin-gonic/gin"
)

func OpenState(st *state.State, group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {
		data := map[string]interface{}{
			"space":     st.Open.Space,
			"radstelle": st.Open.Radstelle,
			"lab3d":     st.Open.Lab3d,
			"machining": st.Open.Machining,
		}

		c.JSON(200, data)
	})
}
