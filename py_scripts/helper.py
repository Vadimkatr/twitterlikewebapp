import requests

def main():
    print("Start script")

    url = "http://localhost:8080"
    user1 = {
        "email": "vadim@gmail.com",
        "username": "Vadim",
        "password": "password_first"
    }
    user2 = {
        "email": "dima@mail.ru",
        "username": "Dima",
        "password": "password_second"
    }
    user3 = {
        "email": "alex@asd.en",
        "username": "Alex",
        "password": "password_third"
    }
    users = [user1, user2, user3]
    # register all users
    for user in users:
        r = requests.post(f"{url}/register", json=user)
    
    # login as user1 and create two tweets
    ru1 = requests.post(f"{url}/login", json={'email': user1['email'], 'password': user1['password']})
    r = requests.post(
        f"{url}/tweets", 
        json={'message': 'Hello, Im Vadim and this is my first tweet'},
        cookies=ru1.cookies
    )
    r = requests.post(
        f"{url}/tweets", 
        json={'message': 'My day is greate day (c) Vadim'},
        cookies=ru1.cookies
    )

    # login as user2, create one tweets and subscribe to first user
    ru2 = requests.post(f"{url}/login", json={'email': user2['email'], 'password': user2['password']})
    r = requests.post(
        f"{url}/tweets", 
        json={'message': 'Im a bad boy (c) Dima'},
        cookies=ru2.cookies
    )
    r = requests.post(
        f"{url}/subscribe", 
        json={'nickname': user1['username']},
        cookies=ru2.cookies
    )

    # login as user3 and subscribe to first and second user
    ru3 = requests.post(f"{url}/login", json={'email': user3['email'], 'password': user3['password']})
    r = requests.post(
        f"{url}/subscribe", 
        json={'nickname': user1['username']},
        cookies=ru3.cookies
    )
    r = requests.post(
        f"{url}/subscribe", 
        json={'nickname': user2['username']},
        cookies=ru3.cookies
    )

if __name__ == '__main__':
    main()
