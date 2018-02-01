// url to make requests too
var url = "http://localhost:8080/";
var session_id;

// url of this website
var website_url = "https://www.hello.co.uk/";

// event listener for window resizes
window.onresize = resize;
var height = window.innerHeight;
var width = window.innerWidth;


function post_website_url(callback) {
  // create xhttp req object
  var request = new XMLHttpRequest();

  // open a post request
  request.open('POST', url);

  // set response type to json
  request.responseType = 'json';

  // define callback upon response
  request.onload = function() {
    callback(request.response)
  };

  // define the request payload
  var data = {
    "WebsiteUrl": website_url,
    "SessionId": session_id
  };

  // finally, make the request
  request.send( JSON.stringify(data) );

}

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
      callback();
    }
  }

  // make the request
  request.send();
}

function post_resize(ResizeFrom, ResizeTo, callback) {
  var request = new XMLHttpRequest();
  var data = {
    "WebsiteUrl": website_url,
    "SessionId": session_id,
    "ResizeFrom": ResizeFrom,
    "ResizeTo": ResizeTo
  };
  request.open('POST', url);
  request.responseType = 'json';
  request.onload = function() {
    callback(request.response);
  };
  request.send( JSON.stringify(data) );
}


function resize() {
  // prepare ResizeFrom field for request
  var ResizeFrom = { "Height": height.toString(), "Width": width.toString() }

  // update width and height
  height = window.innerHeight
  width  = window.innerWidth

  // prepare ResizeTo field for request
  var ResizeTo = { "Height": height.toString(), "Width": width.toString() }

  // make a post request with new and old dimensions
  post_resize(ResizeFrom, ResizeTo, function(response) {
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
    post_website_url(function(response) {
      console.log(response);
    });
  }
}
