package servers

import (
	"github.com/NatthawutSK/ri-shop/modules/monitor/monitorHandlers"
	"github.com/gofiber/fiber/v2"
)

type IModuleFactory interface{
	MonitorModule()
}


type moduleFactory struct {
	r fiber.Router
	s *server
}

func InitModule(r fiber.Router, s *server) IModuleFactory {
	return &moduleFactory{
		r: r,
		s: s,
	}
}

func (m *moduleFactory) MonitorModule() {
	handle := monitorHandlers.MonitorHandler(m.s.cfg)

	m.r.Get("/", handle.HealthCheck)
}