package main

const indexView = `
<!DOCTYPE html>
<html>
<head>
<title>Media Rank</title>
<style>
  .container {
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-gap: 1em;
  }
  img {
    max-width: 100%;
  }
  .image {
    display: flex;
    align-self: center;
    justify-content: center;
  }
  form {
    text-align: center;
  }
  input[type=submit] {
    font-size: larger;
  }
</style>
</head>
<body>
<div class="container">
  <div class="image">
    <img src="/media/{{.Media1.Id}}" title="Id: {{.Media1.Id}}, Score: {{.Media1.Score}}">
  </div>
  <div class="image">
    <img src="/media/{{.Media2.Id}}" title="Id: {{.Media2.Id}}, Score: {{.Media2.Score}}">
  </div>
  <form action="/vote" method="POST">
    <input type="hidden" name="loser" value="{{.Media2.Id}}">
    <input type="hidden" name="winner" value="{{.Media1.Id}}">
    <input type="submit" value="Winner">
  </form>
  <form action="/vote" method="POST">
    <input type="hidden" name="loser" value="{{.Media1.Id}}">
    <input type="hidden" name="winner" value="{{.Media2.Id}}">
    <input type="submit" value="Winner">
  </form>
</div>
</body>
</html>
`

const listView = `
<!DOCTYPE html>
<html>
<head>
<title>Media Rank</title>
<style>
</style>
</head>
<body>
  <table>
  {{range .List}}
    <tr>
    <td><a href="/media/{{.Id}}">{{.Path}}</a></td>
    <td>{{.Score}}</td>
    </tr>
  {{end}}
  </table>
</body>
</html>
`
