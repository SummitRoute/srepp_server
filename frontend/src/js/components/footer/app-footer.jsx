var React = require('react');
var Link = require('react-router-component').Link;

var Footer =
  React.createClass({
    render:function(){
      return (
        <div className="footer">
          <div className="container">
            <p className="text-muted">
              &copy; Summit Route LLC 2015 &mdash;
              <Link href="/privacy_policy">Privacy Policy</Link> &mdash;
              <Link href="/terms_and_conditions">Terms of Use</Link>
            </p>
          </div>
        </div>
      )
    }
  });
module.exports = Footer;
