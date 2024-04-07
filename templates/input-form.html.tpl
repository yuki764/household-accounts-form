<html lang="ja-jp">
<!DOCTYPE html>
<head>
	<title>Household Accounts Input Form</title>
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<meta name="referrer" content="no-referrer">

	<link rel="manifest" href="manifest.webmanifest">

	<style>
		input, select, button {
			font-size: 2em;
			height: 2em;
			width: 20em;
			max-width: 90vw;
			margin: 0.2em;
			padding: 0 8px;
			border-radius: 8px;
			border-style: none;
			background-color: #e5e4e2;
		}
		input:focus, select:focus {
			background-color: #fff;
		}
		button, .submit {
			background-color: #e6e6fa;
			border-style: solid;
			border-width: 1px;
		}
		div.flex {
			display: flex;
			justify-content: center;
			flex-wrap: wrap;
		}
		div.flex input, button {
			display: block;
		}
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

	<script type="text/javascript">
		function sumPrices() {
			let sum = 0;
			[...document.getElementById("price_inputs").getElementsByTagName("input")].forEach((p) => {
				sum += Number(p.value);
			});
			document.getElementById("price").value = sum;
		};

		function appendPriceInput() {
			const newInput = document.createElement("input");
			newInput.setAttribute("type", "number");
			newInput.setAttribute("onchange", "sumPrices()");
			document.getElementById("price_inputs").appendChild(newInput);
		}

		function concatItem() {
			detail = document.getElementById("item_detail").value;
			if (detail != "") {
				detail = ": " + detail
			}
			document.getElementById("item").value = document.getElementById("item_summary").value + detail
		}
	</script>
</head>
<body>
	<form method="post" action="account">
		<input required id="date" name="date" type="date" value="{{.date}}">
		<select required id="category" name="category">
			{{range .categoryList}}
			<option {{if $.submit}}{{if eq . $.submit.category}}selected {{end}}{{end}}value="{{.}}">{{.}}</option>
			{{end}}
		</select>
		<div class="flex">
			<div>
				<div id="price_inputs">
					<input required type="number" onchange="sumPrices()" placeholder="Price">
				</div>
				<button type="button" onclick="appendPriceInput()">+</button>
				<input required id="price" name="price" readonly type="hidden">
			</div>
			<div>
				<input required id="item_summary" onchange="concatItem()" placeholder="Item Summary">
				<input          id="item_detail"  onchange="concatItem()" placeholder="Item Detail">
				<input required id="item" name="item" readonly type="hidden">
			</div>
		</div>
		<input class="submit" type="submit">
	</form>
	{{ with .submit }}
	<ul>
		<li>date: {{ .date }}</li>
		<li>category: {{ .category }}</li>
		<li>price: {{ .price }}</li>
		<li>item: {{ .item }}</li>
	</ul>
	{{ end }}
</body>
</html>
