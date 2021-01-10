<html>
<!DOCTYPE html>
<head>
	<title>Household Accounts Input Form</title>
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<style>
		input, select {font-size: 2em; width: 20em; max-width: 90vw; margin: 0.2em;}
		form {margin: 1em 0;}
		body {text-align: center;}
		ul {text-align: left;}
		@media only screen and (min-width: 768px) {
			input#date {width: 10em;}
			select#category {width: 10em;}
		}
		@media only screen and (max-width: 767px) {
			input, select {width: 95vw;}
		}
	</style>
</head>
<body>
	<form method="post" action="account">
		<input required id="date" name="date" type="date" value="{{.date}}">
		<select required id="category" name="category">
			{{range .categoryList}}
			<option {{if $.submit}}{{if eq . $.submit.category}}selected {{end}}{{end}}value="{{.}}">{{.}}</option>
			{{end}}
		</select>
		<input required id="price" name="price" type="number">
		<input required id="item" name="item">
		<input type="submit">
	</form>
	{{ with .submit }}
	<ul>
		<li>date: {{ .date }}</li>
		<li>category: {{ .category }}</li>
		<li>price: {{ .price }}</li>
		<li>item: {{ .item }}</li>
		<li><a href="/">return</a>
	</ul>
	{{ end }}
</body>
</html>
