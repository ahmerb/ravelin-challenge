// url to make requests too
var url = "http://localhost:8080/";
var session_id;

// url of this website
var website_url = "https://www.hello.co.uk/";

// event listener for window resizes
window.onresize = resize;
var height = window.innerHeight;
var width = window.innerWidth;


// post request-ers

function post_new_session(callback) {
  // create xhttp req object
  var request = new XMLHttpRequest();

  // path
  var session_path = "session"

  // open a post request
  request.open('POST', url + session_path);

  // set response type to json
  request.responseType = 'json';

  // define callback upon release
  request.onload = function() {
    if (request.response) {
      session_id = request.response.SessionId;
      callback(request.response);
    }
  };

  // payload
  var data = {
    "websiteUrl": website_url
  };

  // make the request
  request.send( JSON.stringify(data) );
}

function post_resize(ResizeFrom, ResizeTo, callback) {
  var request = new XMLHttpRequest();
  var data = {
    "eventType": "resizeWindow",
    "websiteUrl": website_url,
    "sessionId": session_id,
    "resizeFrom": ResizeFrom,
    "resizeTo": ResizeTo
  };
  request.open('POST', url);
  request.responseType = 'json';
  request.onload = function() {
    callback(request.response);
  };
  request.send( JSON.stringify(data) );
}

function post_form_elem_copypaste(data, callback) {
  var request = new XMLHttpRequest();
  request.open('POST', url);
  request.responseType = 'json';
  request.onload = function() {
    callback(request.response);
  };
  request.send( JSON.stringify(data) );
}


// event listeners

function resize() {
  // prepare ResizeFrom field for request
  var ResizeFrom = { "height": height.toString(), "width": width.toString() }

  // update width and height
  height = window.innerHeight
  width  = window.innerWidth

  // prepare ResizeTo field for request
  var ResizeTo = { "height": height.toString(), "width": width.toString() }

  // make a post request with new and old dimensions
  post_resize(ResizeFrom, ResizeTo, function(response) {
    console.log(response);
  });
}

function form_elem(pasted, form_id) {
  var pasted = pasted == "paste";

  var data = {
    "eventType": "copyAndPaste",
    "websiteUrl": website_url,
    "sessionId": session_id,
    "pasted": pasted,
    "formId": form_id
  };

  post_form_elem_copypaste(data, function(response) {
    console.log(response);
  });
}




// ========================

main();

function main() {
  // create a new session
  post_new_session(run);
}

function run() {
  if (!session_id && session_id == "") {
    console.log("failed to create session");
  }
  else {
    console.log(session_id);
  }
}
