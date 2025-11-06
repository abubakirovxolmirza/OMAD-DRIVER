package handlers

import "github.com/gofiber/fiber/v2"

// CreateTaxiOrderFiber - Fiber version
func (h *OrderHandler) CreateTaxiOrderFiber(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.H{"error": "Not implemented yet"})
}

// CreateDeliveryOrderFiber - Fiber version
func (h *OrderHandler) CreateDeliveryOrderFiber(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.H{"error": "Not implemented yet"})
}

// GetMyOrdersFiber - Fiber version
func (h *OrderHandler) GetMyOrdersFiber(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.H{"error": "Not implemented yet"})
}

// GetOrderByIDFiber - Fiber version
func (h *OrderHandler) GetOrderByIDFiber(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.H{"error": "Not implemented yet"})
}

// CancelOrderFiber - Fiber version
func (h *OrderHandler) CancelOrderFiber(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.H{"error": "Not implemented yet"})
}
