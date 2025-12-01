package routes

import (
	"github.com/mshubham024/go-url-shortner/database"
	"github.com/redis/go-redis/v9"
	"github.com/gofiber/fiber/v2"
)

//Resolving the redis client from database package

func ResolUrl(c *fiber.Ctx) error{
	url:=c.Params("url")
	rdb:=database.CreateClient(0)
	defer rdb.Close()
	orignalUrl,err:=rdb.Get(database.Ctx,url).Result()
	if err==redis.Nil{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message":"Url not fount on database",
		})
	}else if err!=nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message":"Failed to connect to the database"
		})
	}
	rIncr:=database.CreateClient(1)
	rIncr.Incr(database.Ctx,"Counter")
	defer rIncr.Close()

}