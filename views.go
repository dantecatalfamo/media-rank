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
    border-radius: 4px;
    box-shadow: 0px 1px 2px #0000005e;
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
  .list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    grid-gap: 20px;
    max-width: 1200px;
    margin: auto;
  }
  .list-entry {
    text-align: center;
    display: flex;
    flex-direction: column;
    justify-content: end;
    min-height: 100px;
  }
  .entry-image {
    display: flex;
    justify-content: center;
    align-items: center;
    flex-grow: 1;
  }
  .entry-number {
    font-weight: bold;
  }
  img {
    max-height: 200px;
    max-width: 200px;
    border-radius: 3px;
    box-shadow: 0px 1px 2px #0000005e;
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
  <div class="list">
  {{range $i, $e := .List}}
    <div class="list-entry">
      <div class="entry-image"><a href="/media/{{$e.Id}}" target="_blank"><img title={{.Path}} src="/media/{{$e.Id}}" loading="lazy"></a></div>
      <div class="entry-info">
        <span class="entry-number">{{$i}}</span>
        <span class="entry-score">({{$e.Score}})</span>
      </div>
    </div>
  {{end}}
  </div>
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
  .heading {
    text-align: center;
    margin-bottom: 10px;
    font-weight: bold;
  }
  .list {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    max-width: fit-content;
    margin: auto;
  }
  .winner {
    justify-content: right;
  }
  .score {
    margin: 20px;
    display: flex;
    align-items: center;
  }
  .loser {
    justify-content: left;
  }
  .image {
    display: flex;
    align-items: center;
    margin-bottom: 10px;
    min-height: 100px;
  }
  img {
    max-height: 200px;
    max-width: 200px;
    border-radius: 3px;
    box-shadow: 0px 1px 2px #0000005e;
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
  <div class="list">
  <span class="heading">Winner</span>
  <span class="heading">Points</span>
  <span class="heading">Loser</span>
  {{range .Comparisons}}
    <div class="winner image"><a href="/media/{{.Winner.Id}}" target="_blank"><img src="/media/{{.Winner.Id}}" title="{{.Winner.Path}}" loading="lazy"></a></div>
    <div class="score">{{.Points}}</div>
    <div class="loser image"><a href="/media/{{.Loser.Id}}" target="_blank"><img src="/media/{{.Loser.Id}}" title="{{.Loser.Path}}" loading="lazy"></a></div>
  {{end}}
  </div>
</body>
</html>
`
