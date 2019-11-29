# twitterlikewebapp

## Package describe
internal/app/apiserver/apiserver.go: Contains func "Start" in wich init db connection (I use mysql), init http server (use gorilla/mux as HTTP router) and start http.ListenAndServe.

internal/app/apiserver/server.go: Contains http router with handler funcs.

internal/app/model/*: Describe models that saved in db.

ddl/create_entities.ddl: Describe model tables;use them to create tables and insert values in them with test purpose.

internal/app/store/store.go: Contains Store interface.

internal/app/store/mysqlstore/*: Implement MySQL store with repositoryes.


internal/app/store/teststore/*: Store that use only in test purpose.

## Server routes

### `/register` - POST request with payload
```
{
    "email": "testuser@gmail.com",
    "username": "TestUser",
    "password": "asdzxc"
}
```
Create user in `users` table. Result is:
```
{
    "id": 1,
    "username": "TestUser",
    "email": "testuser@gmail.com"
}
```

### `/login` - POST request with payload
```
{
    "email": "testuser@gmail.com",
    "password": "asdzxc"
}
```
Find user in db. If ok, then create jwt and set them in cookies. Result is:
```
{
    "token": "some json web token",
}
```

### `/tweets` - POST request with payload
```
{
    "message": "some tweet"
}
```
Check user with jwt. If ok, then create user tweet in `tweets` table. Result is:
```
{
    "id": "tweet id",
    "token": "some json web token",
}
```

### `/tweets` - GET request.

Check user with jwt. If ok, then find all tweets of user subscribtions. Result is:
```
{
    "tweets": "["first tweet of publisher", "second", ...]",
}
```

### `/mytweets` - GET request.

Check user with jwt. If ok, then find all user tweets. Result is:
```
{
    "tweets": "["first user tweet", "second", ...]",
}
```

### `/subscribe` - POST request with payload
```
{
    "nickname": "some username"
}
```
Check user with jwt. If ok, then  find nickname from payload in db, if ok, them add note to `subscribers` table that user subscribe to another user. Result is:
