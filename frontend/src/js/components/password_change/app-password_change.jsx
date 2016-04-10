var React = require('react');
var Reqwest = require("reqwest");

var FormInput = require('../form_input.jsx');
var Flash = require('../flash.jsx');
var CSRF = require('../csrf.jsx');

var PasswordChange =
  React.createClass({
    getInitialState: function() {
      return {
        isSaving: false,
        canSave: false,
      }
    },

    handleChange: function() {
      password1 = this.refs.new_password.getInput();
      password2 = this.refs.new_password2.getInput();

      if (password2 != "") {
        if (password2 != password1) {
          this.refs.new_password2.SetBad();
          this.setState({canSave: false});
          return;
        } else {
          this.refs.new_password2.SetGood();
        }
      }

      if (
        this.refs.new_password.getInput() != "" &&
        this.refs.new_password2.getInput() != "" &&
        this.refs.current_password.getInput() != ""
      ) {
        this.setState({canSave: true});
      }
    },

    handleSubmit: function() {
      if (this.state.isSaving == true) { return; }

      this.setState({isSaving: true, canSave: false});

      var thisComponent = this;

      Reqwest({
        url: '/api/change_password.json',
        type: 'json',
        method: 'post',
        headers: {
          'X-CSRF-Token': CSRF()
        },
        data: {
          NewPassword: thisComponent.refs.new_password.getInput(),
          CurrentPassword: thisComponent.refs.current_password.getInput()
        },
        success:function(resp){
          thisComponent.refs.new_password.Reset("");
          thisComponent.refs.new_password2.Reset("");
          thisComponent.refs.current_password.Reset("");
          thisComponent.setState({isSaving: false});

          thisComponent.refs.flash.Show("success", "Password changed");
        },
        error: function (err) {
          msg = "Server error, try again later";
          if (err.response == "old password incorrect") {
            msg = "Current password was incorrect";
          }
          thisComponent.setState({isSaving: false});
          thisComponent.refs.flash.Show("danger", msg);
        }
      })
    },


    isSavingDenied: function() {
      if (this.state.isSaving) {
        return true;
      }
      if (this.state.canSave) {
        return false;
      }
      return true;
    },

    render:function(){
      return (
        <form>
          <Flash ref="flash"/>
          <fieldset>
            <legend>Change Password</legend>

            <FormInput label="New password" type="password" defaultVal="" whenChanged={this.handleChange} ref="new_password" addon="password" />
            <FormInput label="Re-enter new password" type="password" defaultVal="" whenChanged={this.handleChange} ref="new_password2" addon="password"/>
            <FormInput label="Current Password" type="password" defaultVal="" whenChanged={this.handleChange} ref="current_password" addon="password"/>

            <button id="changePassword" name="changePassword" className="btn btn-primary"
              disabled={this.isSavingDenied()}
              onClick={!this.state.isSaving ? this.handleSubmit : null}>
              {this.state.isSaving ? 'Changing...' : 'Change Password'}</button>

          </fieldset>

        </form>
        )
    }
  });
module.exports = PasswordChange;
