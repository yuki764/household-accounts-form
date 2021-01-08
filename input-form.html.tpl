<html>
<head>
	<title>Household Accounts Input Form</title>
	<style>
		input, select {font-size: 2em; width: 80vw; margin: 0.2em;}
		form {margin: 1em;}
		body {text-align: center;}
		ul {text-align: left;}
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
