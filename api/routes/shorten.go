package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/mshubham024/go-url-shortner/database"
	"github.com/mshubham024/go-url-shortner/helpers"
	"github.com/redis/go-redis/v9"
	"github.com/teris-io/shortid"
)

type request struct {
	Url         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	Url             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot parse JSON",
		})
	}
	rdb2 := database.CreateClient(1)
	defer rdb2.Close()
	_, err := rdb2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		_ = rdb2.Set(database.Ctx, c.IP(), os.Getenv("API_QOUTA"), time.Minute*30)
	} else {
		val, _ := rdb2.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := rdb2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"message":          "Rate limit exceeded",
				"rate_limit_reset": limit.Minutes(),
			})
		}
	}
	//validate url
	if !govalidator.IsURL(body.Url) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Not valid Url",
		})
	}
	if !helpers.RemoveDomainError(body.Url) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"message": "You cannot shorten the domain url",
		})
	}
	body.Url = helpers.EnforceHTTP(body.Url)

	//creating short url
	var id string
	if body.CustomShort == "" {
		id = shortid.MustGenerate()
	} else {
		id = body.CustomShort
	}
	rdb := database.CreateClient(0)
	defer rdb.Close()
	val, _ := rdb.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Already have these shorthen url",
		})
	}
	if body.Expiry == 0 {
		body.Expiry = 24 * time.Hour
	}
	err = rdb.Set(database.Ctx, id, body.Url, body.Expiry).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not connect to the server",
		})
	}
	resp := response{
		Url:             body.Url,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}
	rdb2.Decr(database.Ctx, c.IP())
	val, _ = rdb2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)
	ttl, _ := rdb2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = time.Duration(ttl.Minutes()) * time.Minute
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)

}
