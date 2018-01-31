// url to make requests too
var url = "http://localhost:8080/";

// create xhttp req object
var request = new XMLHttpRequest();

// open a post request
request.open('POST', url);

// set response type to json
request.responseType = 'json';

// define callback upon response
request.onload = function() {
  console.log(request.response);
  console.log(request);
}

// define the request payload
var data = {
  "SessionId": "1234567890"
};

// finally, make the request
request.send( JSON.stringify(data) );
