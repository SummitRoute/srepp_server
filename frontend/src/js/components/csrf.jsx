var CSRF = function() {
  var csrf = "";
  var metas = document.getElementsByTagName('meta');

  for (i=0; i<metas.length; i++) {
    if (metas[i].getAttribute("name") == "csrf-token") {
      csrf = metas[i].getAttribute("content");
    }
  }
  return csrf;
};
module.exports = CSRF;
