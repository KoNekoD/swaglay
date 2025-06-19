package swaglay_fiber

import (
	"github.com/gofiber/fiber/v3"
	"regexp"
)

func replacePath(path string) string {
	return regexp.MustCompile(`\{(\S+)\}`).ReplaceAllString(path, ":$1")
}

func fullPath(path string) string {
	if v, ok := Fiber.(*fiber.Group); ok {
		return v.Prefix + path
	}

	return path
}
