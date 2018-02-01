// url to make requests too
var url = "http://localhost:8080/";
var session_id;


function post_website_url(website_url, callback) {
  // create xhttp req object
  var request = new XMLHttpRequest();

  // open a post request
  request.open('POST', url);

  // set response type to json
  request.responseType = 'json';

  // define callback upon response
  request.onload = function() {
    callback(request.response)
  }

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


// create a new session
post_new_session(run);

function run() {
  if (!session_id && session_id == "") {
    console.log("failed to create session");
  }
  else {
    console.log(session_id);
    post_website_url("hello.co.uk", function(response) {
      console.log(response);
    });
  }
}
