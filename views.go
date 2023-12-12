package main

const indexView = `
<!DOCTYPE html>
<html>
<head>
<title>Media Rank</title>
<style>
  html {
    font-family: "Open Sans", "Helvetica", "sans";
  }
  .selection {
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-gap: 1em;
  }
  img {
    max-width: 100%;
    max-height: 80vh;
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
  h1 {
    margin-bottom: 10px;
  }
  header {
    text-align: center;
    margin-bottom: 40px;
  }
</style>
</head>
<body>
  <header>
    <h1>Media Rank</h1>
    <div><a href="/list">Ranked List</a> <a href="/history">History</a></div>
  </header>
  <div class="selection">
    <div class="image">
      <a href="/media/{{.Media1.Id}}" target="_blank">
        <img src="/media/{{.Media1.Id}}" title="Id: {{.Media1.Id}}, Score: {{.Media1.Score}}, Path: {{.Media1.Path}}">
      </a>
    </div>
    <div class="image">
      <a href="/media/{{.Media2.Id}}" target="_blank">
        <img src="/media/{{.Media2.Id}}" title="Id: {{.Media2.Id}}, Score: {{.Media2.Score}}, Path: {{.Media2.Path}}">
      </a>
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
  html {
    font-family: "Open Sans", "Helvetica", "sans";
  }
  table {
    margin: auto;
  }
  .table-number, .table-score {
    text-align: right;
  }
  .table-number {
    font-weight: bold;
  }
  .table-score {
    padding-right: 15px;
    padding-left: 15px;
  }
  th {
    border-bottom: 1px solid black;
  }
  img {
    max-height: 100px;
    max-width: 100px;
  }
  header {
    text-align: center;
    margin-bottom: 40px;
  }
</style>
</head>
<body>
  <header>
    <h1>Media Rank</h1>
    <div><a href="/">Face Off</a> <a href="/history">History</a></div>
  </header>
  <table>
  <tr>
    <th>Rank</th>
    <th>Score</th>
    <th>Preview</th>
  </tr>
  {{range $i, $e := .List}}
    <tr>
      <td class="table-number">{{$i}}</td>
      <td class="table-score">{{$e.Score}}</td>
      <td class="table-image"><a href="/media/{{$e.Id}}" target="_blank"><img title={{.Path}} src="/media/{{$e.Id}}" loading="lazy"></a></td>
    </tr>
  {{end}}
  </table>
</body>
</html>
`

const historyTmpl = `
<!DOCTYPE html>
<html>
<head>
<title>Media Rank</title>
<style>
  html {
    font-family: "Open Sans", "Helvetica", "sans";
  }
  table {
    margin: auto;
  }
  .table-score {
    padding-right: 15px;
    padding-left: 15px;
    text-align: center;
  }
  th {
    border-bottom: 1px solid black;
  }
  img {
    max-height: 100px;
    max-width: 100px;
  }
  header {
    text-align: center;
    margin-bottom: 40px;
  }
</style>
</head>
<body>
  <header>
    <h1>Media Rank</h1>
    <div><a href="/">Face Off</a> <a href="/list">Ranked List</a></div>
  </header>
  <table>
  <tr>
    <th>Winner</th>
    <th>Points</th>
    <th>Loser</th>
  </tr>
  {{range .Comparisons}}
    <tr>
      <td><a href="/media/{{.Winner.Id}}" target="_blank"><img src="/media/{{.Winner.Id}}" title="{{.Winner.Path}}" loading="lazy"></a></td>
      <td class="table-score">{{.Points}}</td>
      <td><a href="/media/{{.Loser.Id}}" target="_blank"><img src="/media/{{.Loser.Id}}" title="{{.Loser.Path}}" loading="lazy"></a></td>
    </tr>
  {{end}}
  </table>
</body>
</html>
`
