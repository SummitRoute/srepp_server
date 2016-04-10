var React = require('react');

var Header = require('./header/app-header.jsx');
var Footer = require('./footer/app-footer.jsx');

var Template =
  React.createClass({
    render:function(){
      return  (
        <div>
          <Header />
          <div className="container">
            {this.props.children}
          </div>
          <Footer />
        </div>
        )
    }
  });
module.exports = Template;
