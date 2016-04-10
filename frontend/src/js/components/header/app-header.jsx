var React = require('react');
var Link = require('react-router-component').Link;

var Header =
  React.createClass({
    render:function(){
      return (
        <div className="navbar navbar-default" role="navigation">
          <div className="container">
            <div className="navbar-header">
              <button type="button" className="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse">
                <span className="sr-only">Toggle navigation</span>
                <span className="icon-bar"></span>
                <span className="icon-bar"></span>
                <span className="icon-bar"></span>
              </button>
              <Link className="navbar-brand" href="/"></Link>
            </div>

            <div className="navbar-collapse collapse">
              <ul className="nav navbar-nav">
                <li><Link href="/"><span className="glyphicon glyphicon-home"></span> Home</Link></li>
                <li><Link href="/systems">Systems</Link></li>
                <li><Link href="/executables">Executables</Link></li>
              </ul>

              <ul className="nav navbar-nav navbar-right">
                <li>
                  <p className="navbar-text">
                    <Link href="/profile"><span className="glyphicon glyphicon-user"></span>Profile</Link>
                  </p>
                </li>
                <li><a href="/logout">Logout</a></li>
                <li><Link href="/help">Help</Link></li>
              </ul>

            </div>
          </div>
        </div>
        )
    }
  });
module.exports = Header;
