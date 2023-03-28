function get_ui_alert_check_in_allowed() {
	return `
	<div class="row">
		<div class="col-md-4"></div>
		<div class="col-md-4">
			<div class="alert alert-success" id="checked-in-alert-true">
				<center>Allowed to Check In !!!</center>
			</div>
		</div>
		<div class="col-md-4"></div>
	</div>`;
}

function get_ui_alert_check_in_failed() {
	return `
	<div class="row">
		<div class="col-md-3"></div>
		<div class="col-md-6">
			<div class="alert alert-danger" id="checked-in-alert-false">
				<center>
					Checked In Too Recently !!!<br><br>
					<a id="block-button" class="btn btn-warning" target="_blank" href="/none">Block</a>
				</center>
			</div>
		</div>
		<div class="col-md-3"></div>
	</div>`;
}

function get_ui_active_user_info() {
	return `
	<div class="row">
		<center><h2 id="active-username"></h2></center>
		<center><h4 id="active-user-time-remaining"></h4></center>
	</row>
	`;
}

function get_ui_shopping_for_selector() {
	return `
	<div class="row">
		<div class="col-md-3"></div>
		<div class="col-md-6">
			<div class="input-group">
				<div class="input-group-text">Shopping For</div>
				<select id="shopping_for" class="form-select" aria-label="Shopping For" name="shopping_for">
					<option value="1">1</option>
					<option value="2">2</option>
					<option value="3">3</option>
					<option value="4">4</option>
					<option value="5">5</option>
					<option value="6">6</option>
					<option value="7">7</option>
				</select>
			</div>
		</div>
		<div class="col-md-3"></div>
	</div>
	`;
}

function get_ui_user_search_table() {
	return `
	<div class="row">
		<div class="col-md-1"></div>
		<div class="col-md-10">
			<div class="table-responsive-sm">
				<table id="user-search-table" class="table table-hover table-striped-columns">
					<thead>
						<tr>
							<th scope="col">#</th>
							<th scope="col">Username</th>
							<th scope="col">UUID</th>
							<th scope="col">Select</th>
						</tr>
					</thead>
					<tbody id="user-search-table-body"></tbody>
				</table>
			</div>
		</div>
		<div class="col-md-1"></div>
	</div>`;
}
function populate_user_search_table( users ) {
	// console.log( "populate_user_search_table()" );
	// console.log( users );
	$( "#user-search-table" ).show();
	let table_body_element = document.getElementById( "user-search-table-body" );
	table_body_element.innerHTML = "";
	for ( let i = 0; i < users.length; ++i ) {
		let _tr = document.createElement( "tr" );

		let user_number = document.createElement( "th" );
		user_number.setAttribute( "scope" , "row" );
		user_number.textContent = `${(i + 1)}`;
		_tr.appendChild( user_number );

		let username = document.createElement( "td" );
		username.textContent = users[ i ][ "username" ];
		_tr.appendChild( username );

		let uuid_holder = document.createElement( "td" );
		let uuid_text = document.createElement( "span" );
		uuid_text.textContent = users[ i ][ "uuid" ];
		uuid_text.innerHTML += "&nbsp;&nbsp;"
		uuid_holder.appendChild( uuid_text );
		_tr.appendChild( uuid_holder );

		let select_button_holder = document.createElement( "td" );
		let select_button = document.createElement( "button" );
		select_button.textContent = "Select"
		select_button.className = "btn btn-success btn-sm";
		select_button.onclick = function() {
			$( "#user-search-input" ).val( users[ i ][ "uuid" ] );
			// $( "#user-search-table" ).hide();
			// check_in_uuid_input();
			// search_input();
			window.USER = users[ i ];
			// _on_check_in_input_change( users[ i ][ "uuid" ] );
			// $( "#main-row" ).trigger( "render_active_user" , users[ i ] );
			window.UI.render_active_user();
		};
		select_button_holder.appendChild( select_button );
		_tr.appendChild( select_button_holder );

		table_body_element.appendChild( _tr );
	}
}

function get_ui_user_balance_table() {
	return `
	<div class="row">
		<div class="col-md-1"></div>
		<div class="col-md-10">
			<div class="table-responsive-sm">
				<table id="user-balance-table" class="table table-hover table-striped-columns">
					<thead>
						<tr>
							<th scope="col">Item</th>
							<th scope="col">Available</th>
							<th scope="col">Limit</th>
							<th scope="col">Total Used</th>
						</tr>
					</thead>
					<tbody id="user-balance-table-body"></tbody>
				</table>
			</div>
		</div>
		<div class="col-md-1"></div>
		<center>
			<button id="print-checkin-button" type="submit" class="btn btn-success">Print</button>
		</center>
	</div>

	`;

}
function _add_balance_row( table_body_element , name , available , limit , used ) {
	let _tr = document.createElement( "tr" );
	let item = document.createElement( "th" );
	item.textContent = name;
	_tr.appendChild( item );
	let _available = document.createElement( "td" );
	let available_input = document.createElement( "input" );
	available_input.setAttribute( "type" , "text" );
	available_input.className = "form-control";
	available_input.value = available;
	available_input.setAttribute( "id" , `balance_${name.toLowerCase()}_available` );
	_available.appendChild( available_input );
	_tr.appendChild( _available );
	let _limit = document.createElement( "td" );
	let limit_input = document.createElement( "input" );
	limit_input.setAttribute( "type" , "text" );
	limit_input.className = "form-control";
	limit_input.value = limit;
	limit_input.setAttribute( "id" , `balance_${name.toLowerCase()}_limit` );
	limit_input.setAttribute( "readonly" , "" );
	_limit.appendChild( limit_input );
	_tr.appendChild( _limit );
	let _used = document.createElement( "td" );
	let used_input = document.createElement( "input" );
	used_input.setAttribute( "type" , "text" );
	used_input.className = "form-control";
	used_input.value = used;
	used_input.setAttribute( "id" , `balance_${name.toLowerCase()}_used` );
	used_input.setAttribute( "readonly" , "" );
	_used.appendChild( used_input );
	_tr.appendChild( _used );
	table_body_element.appendChild( _tr );
}

// could just switch to multiple inputs ?
// https://getbootstrap.com/docs/5.3/forms/input-group/#multiple-inputs
function populate_user_balance_table( shopping_for , balance , balance_config ) {

	console.log( "populate_user_balance_table()" );
	console.log( "shopping for === " , shopping_for );
	console.log( "balance === " , balance );
	console.log( "balance_config === " , balance_config );

	let tops_available = ( shopping_for * balance_config.general.tops );
	let bottoms_available = ( shopping_for * balance_config.general.bottoms );
	let dresses_available = ( shopping_for * balance_config.general.dresses );
	let shoes_available = ( shopping_for * balance_config.shoes );
	let seasonal_available = ( shopping_for * balance_config.seasonals );
	let accessories_available = ( shopping_for * balance_config.accessories );

	let table_body_element = document.getElementById( "user-balance-table-body" );
	table_body_element.innerHTML = "";

	_add_balance_row( table_body_element , "Tops" ,
		tops_available ,
		balance[ "general" ][ "tops" ][ "limit" ] ,
		balance[ "general" ][ "tops" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Bottoms" ,
		bottoms_available ,
		balance[ "general" ][ "bottoms" ][ "limit" ] ,
		balance[ "general" ][ "bottoms" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Dresses" ,
		dresses_available ,
		balance[ "general" ][ "dresses" ][ "limit" ] ,
		balance[ "general" ][ "dresses" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Shoes" ,
		shoes_available ,
		balance[ "shoes" ][ "limit" ] ,
		balance[ "shoes" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Seasonals" ,
		seasonal_available ,
		balance[ "seasonals" ][ "limit" ] ,
		balance[ "seasonals" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Accessories" ,
		accessories_available ,
		balance[ "accessories" ][ "limit" ] ,
		balance[ "accessories" ][ "used" ] ,
	);

}

function get_ui_user_edit_form() {
	return `
	<div class="row">
		<center>
			<form id="user-edit-form" action="/admin/user/edit" onSubmit="return on_submit( event )" method="post">
				<!-- Main Required Stuff -->
				<div class="row g-2 mb-3">
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_first_name" type="text" class="form-control" name="user_first_name">
							<label for="user_first_name">First Name</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_middle_name" type="text" class="form-control" name="user_middle_name">
							<label for="user_middle_name">Middle Name</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_last_name" type="text" class="form-control" name="user_last_name">
							<label for="user_last_name">Last Name</label>
						</div>
					</div>
				</div>
				<div class="row g-2 mb-3">
						<div class="col-md-2"></div>
						<div class="col-md-4">
							<div class="form-floating">
								<input id="user_email" type="email" class="form-control" name="user_email">
								<label for="user_email">Email Address</label>
							</div>
						</div>
						<div class="col-md-4">
							<div class="form-floating">
								<input id="user_phone_number" type="tel" class="form-control" name="user_phone_number">
								<label for="user_phone_number">Phone Number</label>
							</div>
						</div>
						<div class="col-md-2"></div>
				</div>

				<div class="row g-2 mb-3">
					<div class="col-md-4"></div>
					<div class="col-md-4">
						<button id="add-barcode-button" class="btn btn-primary" onclick="on_add_barcode(event);">Add Barcode</button>
					</div>
					<div class="col-md-4"></div>
				</div>

				<div id="user_barcodes"></div>

				<br>

				<!-- Address - Part 1-->
				<div class="row g-2 mb-3">
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_street_number" type="text" class="form-control" name="user_street_number">
							<label for="user_street_number">Street Number</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_street_name" type="text" class="form-control" name="user_street_name">
							<label for="user_street_name">Street Name</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_address_two" type="text" class="form-control" name="user_street_name">
							<label for="user_address_two">Address 2</label>
						</div>
					</div>
				</div>
				<!-- Address - Part 2-->
				<div class="row g-2 mb-3">
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_city" type="text" class="form-control" name="user_city">
							<label for="user_city">City</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_state" type="text" class="form-control" name="user_state">
							<label for="user_state">State</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_zip_code" type="text" class="form-control" name="user_zip_code">
							<label for="user_zip_code">Zip Code</label>
						</div>
					</div>
				</div>
				<br>
				<!-- Extras -->
				<div class="row g-2 mb-3">

					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_birth_day" type="number" min="1" max="31" class="form-control" name="user_birth_day">
							<label for="user_birth_day">Birth Day</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<select id="user_birth_month" class="form-select" aria-label="User Birth Month" name="user_birth_month">
								<option value="JAN">JAN = 1</option>
								<option value="FEB">FEB = 2</option>
								<option value="MAR">MAR = 3</option>
								<option value="APR">APR = 4</option>
								<option value="MAY">MAY = 5</option>
								<option value="JUN">JUN = 6</option>
								<option value="JUL">JUL = 7</option>
								<option value="AUG">AUG = 8</option>
								<option value="SEP">SEP = 9</option>
								<option value="OCT">OCT = 10</option>
								<option value="NOV">NOV = 11</option>
								<option value="DEC">DEC = 12</option>
							</select>
							<label for="user_birth_month">Birth Month</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_birth_year" type="number" min="1900" max="2100" class="form-control" name="user_birth_year">
							<label for="user_birth_year">Birth Year</label>
						</div>
					</div>
				</div>

				<br>

				<div class="row g-2 mb-3">
					<div class="col-md-4"></div>
					<div class="col-md-4">
						<div class="form-floating">
							<select id="user_family_size" class="form-select" aria-label="User Family Size" name="user_family_size">
								<option value="0">0</option>
								<option value="1">1</option>
								<option value="2">2</option>
								<option value="3">3</option>
								<option value="4">4</option>
								<option value="5">5</option>
								<option value="6">6</option>
							</select>
							<label for="user_family_size">Family Members</label>
						</div>
					</div>
					<div class="col-md-4"></div>
				</div>

				<div id="user_family_members">

				</div>

				<br>

				<div class="form-row">
					<button id ="save-button" type="submit" class="btn btn-success">Save</button>
				</div>

			</form>
		</center>
	</div>`;
}
