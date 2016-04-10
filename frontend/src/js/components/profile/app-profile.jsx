var React = require('react');
var Reqwest = require("reqwest");
var Link = require('react-router-component').Link;

var FormInput = require('../form_input.jsx');
var Flash = require('../flash.jsx');
var CSRF = require('../csrf.jsx');

var Profile =
  React.createClass({
    getInitialState: function() {
      return {
        receivedJson: false,
        isSaving: false,
        hasChanges: false,

        // Form vars
        email: "",
        firstname: "",
        lastname: ""
      }
    },

    handleChange: function() {
      this.setState({hasChanges: true});
    },

    handleSubmit: function() {
      if (this.state.isSaving == true) { return; }

      this.setState({isSaving: true, hasChanges: false});

      var thisComponent = this;

      Reqwest({
        url: '/api/profile.json',
        type: 'json',
        method: 'post',
        headers: {
          'X-CSRF-Token': CSRF()
        },
        data: {
          FirstName: thisComponent.refs.firstname.getInput(),
          LastName: thisComponent.refs.lastname.getInput(),
          Email: thisComponent.refs.email.getInput(),
        },
        success:function(resp){
          thisComponent.refs.firstname.Reset();
          thisComponent.refs.lastname.Reset();
          thisComponent.refs.email.Reset();
          thisComponent.setState({isSaving: false});

          thisComponent.refs.flash.Show("success", "Changes saved");
        },
        error: function (err) {
          msg = "Server error, try again later";
          if (err.response == "email not unique") {
            msg = "Account already exists. Please choose a different email address.";
            thisComponent.refs.email.SetBad();
          }
          thisComponent.setState({isSaving: false});
          thisComponent.refs.flash.Show("danger", msg);
        }
      })
    },

    componentWillMount: function() {
      thisComponent = this;
      Reqwest({
        url: '/api/profile.json',
        type: 'json',
        success:function(resp){
          thisComponent.setState({receivedJson: true, firstname: resp.FirstName, lastname: resp.LastName, email: resp.Email});
        },
        error: function (err) {
          thisComponent.refs.flash.Show("danger", "The server is experiencing problems right now");
        }
      })
    },

    isSavingDenied: function() {
      if (this.state.isSaving) {
        return true;
      }
      if (this.state.hasChanges) {
        return false;
      }
      return true;
    },

    render:function(){
      // What to display when no data has been received yet
      if (this.state.receivedJson === false) {
        return (
            <Flash ref="flash"/>
        );
      }

      return (
          <div>
            <Flash ref="flash"/>
            <form>
              <fieldset>
                <legend>Profile</legend>

                <div className="form-group">
                  <FormInput label="First name" defaultVal={this.state.firstname} whenChanged={this.handleChange} ref="firstname" />
                  <FormInput label="Last name" defaultVal={this.state.lastname} whenChanged={this.handleChange} ref="lastname" />
                </div>

                <div className="form-group">
                  <FormInput label="Email" defaultVal={this.state.email} whenChanged={this.handleChange} ref="email" addon="email" />
                </div>

                <div className="form-group">
                  <label className="col-md-4 control-label"></label>
                  <div className="col-md-4">
                    <button id="save" name="save" className="btn btn-primary"
                      disabled={this.isSavingDenied()}
                      onClick={!this.state.isSaving ? this.handleSubmit : null}>
                      {this.state.isSaving ? 'Saving...' : 'Save'}</button>
                  </div>
                  <div className="col-md-4">
                    <Link href="/change_password">Change password</Link>
                  </div>
                </div>

              </fieldset>
            </form>
          </div>
        )
    }
  });
module.exports = Profile;
