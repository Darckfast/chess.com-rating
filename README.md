![logo](.github/images/chess_com.png)

# Chess.com rating
API to get user ratings on chess.com, and returns it as plain text

## Getting Started

### StreamElements

Add a custom command using the ${customapi.url}, like the example bellow

```sh
${customapi.https://darckfast.com/api/chess?username=chess_com_username&message=$(queryescape "current rating on rapid: =rapid and puzzles: =tactics")}
```
### NightBot

Add a custom command using the $(urlfetch url), like the example bellow

```sh
$(urlfetch https://darckfast.com/api/chess?username=chess_com_username&message=$(querystring "current rating on rapid: =rapid and puzzles: =tactics"))
```

### FossaBot

Add a custom command using the $(customapi url), like the example bellow

```sh
$(customapi https://darckfast.com/api/chess?username=chess_com_username&message=$(querystring "current rating on rapid: =rapid and puzzles: =tactics"))
```

### MooBot

Create a URL-Fetch command, like the example bellow

```sh
https://darckfast.com/api/chess?username=chess_com_username&message="current rating on rapid: =rapid and puzzles: =tactics"
```

### cURL

The API can be called directly with a GET HTTP request

```sh
curl -X GET 'https://darckfast.com/api/chess' \
  --url-query 'username=MagnusCarlsen' \
  --url-query 'message=Magnus Carlsen current rating on rapid: =rapid, bullet: =bullet, blitz: =lightning'
```

---

[Check the complete documentation for more](https://darckfast.com/docs/chess)
