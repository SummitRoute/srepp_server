var React = require('react');
var Alert = require('react-bootstrap').Alert;

var Flash = React.createClass({
  getInitialState:function(){
    return {
      style:"flash",
      alertType:"danger",
      msg:"Error"
    }
  },

  render: function() {
      return (
        <div ref="flash" className={this.state.style}>
          <Alert bsStyle={this.state.alertType} onDismiss={this.handleAlertDismiss}>
            {this.state.msg}
          </Alert>
        </div>
      );
  },


  handleAlertDismiss: function() {
    this.setState({style: 'flash'});
  },

  Show: function(alertType, msg) {
    this.setState({style: 'flash show', alertType:alertType, msg:msg});
  }
});
module.exports = Flash;
