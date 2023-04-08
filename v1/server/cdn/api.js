function check_in_uuid( uuid , balance_form_data ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let json_balance_form_data = JSON.stringify( balance_form_data );
			let check_in_url = `/admin/user/checkin/${uuid}`;
			let check_in_response = await fetch( check_in_url , {
				method: "POST" ,
				body: json_balance_form_data
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function check_in_uuid_test( uuid ) {
	return new Promise( async function( resolve , reject ) {
		try {
			// let check_in_url = `/admin/user/checkin/test/${uuid}`;
			let check_in_url = `/admin/user/checkin/test/${uuid}`;
			let check_in_response = await fetch( check_in_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" }
			});
			let response_json = await check_in_response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function get_user_from_barcode( barcode ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_url = `/admin/user/get/barcode/${barcode}`;
			let check_in_response = await fetch( check_in_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" }
			});
			let response_json = await check_in_response.json();
			let user = response_json[ "result" ];
			resolve( user );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function get_user_from_uuid( uuid ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_url = `/admin/user/get/${uuid}`;
			let check_in_response = await fetch( check_in_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" }
			});
			let response_json = await check_in_response.json();
			let user = response_json[ "result" ];
			resolve( user );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function search_username( username ) {
	return new Promise( async function( resolve , reject ) {
		try {
			if ( !username ) { resolve( false ); return; }
			let search_url = `/admin/user/search/username/${username}`;
			let check_in_response = await fetch( search_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			if ( result === "not found" ) { result = false; }
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function fuzzy_search_username( username ) {
	return new Promise( async function( resolve , reject ) {
		try {
			if ( !username ) { resolve( false ); return; }
			let search_url = `/admin/user/search/username/fuzzy/${username}`;
			let check_in_response = await fetch( search_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_edit_user( user_info ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response = await fetch( `/admin/user/edit` , {
				method: "POST" ,
				body: JSON.stringify( user_info )
			});
			let response_json = await response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_new_user( user_info ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response = await fetch( `/admin/user/new` , {
				method: "POST" ,
				body: JSON.stringify( user_info )
			});
			let response_json = await response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_delete_user( uuid ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response = await fetch( `/admin/user/delete/${uuid}` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" }
			});
			let response_json = await response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}