var React = require('react');
var Input = require('react-bootstrap').Input;

var FormInput = React.createClass({
  propTypes: {
    addon: React.PropTypes.string,
    type: React.PropTypes.string,
    label: React.PropTypes.string.isRequired,
    defaultVal: React.PropTypes.string.isRequired,
    whenChanged: React.PropTypes.func
  },

  getDefaultProps: function() {
    return {
      addon: "",
      type: "text"
    };
  },

  getInitialState: function() {
    return {
      hasChanged: false,
      isBad: false,
      value: this.props.defaultVal
    };
  },

  Reset: function(val) {
    if (val != undefined) {
      this.setState({value: val});
    }
    this.setState({hasChanged: false});
  },

  SetBad: function() {
    this.setState({hasChanged: true, isBad: true});
  },

  SetGood: function() {
    this.setState({hasChanged: true, isBad: false});
  },

  validationState: function() {
    if (this.state.isBad) {
      return 'error'
    } else if (this.state.hasChanged) {
      var length = this.state.value.length;
      if (length > 0) return 'success';
    }
    // Else return nothing
    return;
  },

  handleChange: function() {
    this.setState({
      value: this.refs.input.getValue(),
      hasChanged: true,
      isBad: false,
    });
    if (this.props.whenChanged != null) {
      this.props.whenChanged();
    }
  },

  getInput: function() {
    return this.refs.input.getValue();
  },


  render: function() {
    var addonComponent = null;
    if (this.props.addon == "email") {
      addonComponent = <i  className="fa fa-envelope"/>;
    } else if (this.props.addon.indexOf("password") > -1) {
      addonComponent = <i  className="fa fa-key"/>;
    }

    return (
        <Input
          type={this.props.type}
          value={this.state.value}
          label={this.props.label}
          bsStyle={this.validationState()}
          ref="input"
          onChange={this.handleChange} addonBefore={addonComponent}/>
    );
  }
});
module.exports = FormInput;
