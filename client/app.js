// url to make requests too
var url = "http://localhost:8080/";
var session_id;

// url of this website
var website_url = document.location.origin;

// event listener for window resizes
window.onresize = resize;
var height = window.innerHeight;
var width = window.innerWidth;

// track time between first key down and form submit
var keydown = {
  "key_pressed": false,
  "time_start": null,   // time on first key press in an input field (milliseconds)
  "time_end": null,     // time on form submit (milliseconds)
};

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

function post_time_taken(time_taken, callback) {
  var request = new XMLHttpRequest();
  request.open('POST', url);
  request.responseType = 'json';
  request.onload = function() {
    callback(request.response);
  };

  var data = {
    "eventType": "timeTaken",
    "websiteUrl": website_url,
    "sessionId": session_id,
    "time": time_taken
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

function key_down() {
  if (keydown.key_pressed) {
    return;
  }
  var date = new Date();
  keydown.time_start = date.getTime();
  keydown.key_pressed = true;
}

function form_submit() {
  // send timeTakem
  var date = new Date();
  keydown.time_end = date.getTime();
  var time_taken = Math.round((keydown.time_end - keydown.time_start) / 1000);
  post_time_taken(time_taken, function(response) {
    console.log(response);
  });
}

// ========================

main();

function main() {
  // create a new session
  post_new_session(run);
}

function run(response) {
  if (!session_id && session_id == "") {
    console.log("failed to create session");
  }
  else {
    console.log(session_id);

    // set the sessionId in the field
    var form_field_sid = document.getElementById("sessionId");
    form_field_sid.value = session_id;
  }
}
