var React = require('react');
var Reqwest = require("reqwest");
var QS = require('querystring');

var FormInput = require('../form_input.jsx');
var Flash = require('../flash.jsx');
var CSRF = require('../csrf.jsx');

var OverlayMixin = require('react-bootstrap').OverlayMixin;
var Button = require('react-bootstrap').Button;
var Modal = require('../Modal.jsx');


var SystemInfo =
  React.createClass({
    mixins: [OverlayMixin],

    getInitialState: function() {
      return {
        loaded: false,

        isEdit: false,
        isSaving: false,
        hasChanges: false,

        isModalOpen: false,

        UUID: '',
        AgentVersion: '',
        OS: '',
        Manufacturer: '',
        Model: '',
        LastSeen: '',
        FirstSeen: '',
        MachineGUID: '',

        Comment: '',
      };
    },

    handleToggle: function () {
      this.setState({
        isModalOpen: !this.state.isModalOpen
      });
    },

    handleChange: function() {
      this.setState({hasChanges: true});
    },

    handleSave: function(e) {
      if (this.state.isSaving == true) { return; }

      this.setState({isSaving: true, hasChanges: false});

      var thisComponent = this;

      Reqwest({
        url: '/api/systeminfo.json',
        type: 'json',
        method: 'post',
        headers: {
          'X-CSRF-Token': CSRF()
        },
        data: {
          Comment: thisComponent.refs.comment.getInput(),
        },
        success:function(resp){
          thisComponent.refs.comment.Reset();
          thisComponent.setState({isSaving: false});

          thisComponent.refs.flash.Show("success", "Changes saved");
        },
        error: function (err) {
          msg = "Server error, try again later";
          thisComponent.setState({isSaving: false});
          thisComponent.refs.flash.Show("danger", msg);
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

    handleForget: function() {
      // Called when the user clicks the button to forget an agent
      this.handleToggle();
      return;
    },

    componentDidMount:function(){
      var thisComponent = this;
      var query = QS.parse(window.location.search.slice(1));

      Reqwest({
        url: '/api/systeminfo.json?uuid='+query["uuid"],
        type: 'json',
        success: function (resp) {
          if (thisComponent.isMounted()) {
            thisComponent.setState({
             UUID: resp.System,
             MachineGUID: resp.MachineGUID,
             AgentVersion: resp.AgentVersion,
             Comment: resp.Comment,
             OS: resp.OSHumanName,
             Manufacturer: resp.Manufacturer,
             Model: resp.Model,
             LastSeen: resp.LastSeen,
             FirstSeen: resp.FirstSeen
            });
          }
        }
      });
    },


    showUserData: function() {
      if (this.state.isEdit) {
        return (
          <fieldset>
            <legend>User defined data</legend>
            <div className="form-group">
              <table>
                <tr><td className="datalabel">Comment</td><td><FormInput defaultVal={this.state.Comment} whenChanged={this.handleChange} ref="comment" /></td></tr>
              </table>
            </div>

            <div className="form-group">
              <div className="col-md-6">
                <Button bsStyle="primary"
                  disabled={this.isSavingDenied()}
                  onClick={!this.state.isSaving ? this.handleSave : null}>
                  {this.state.isSaving ? 'Saving...' : 'Save'}</Button>
              </div>
              <div className="col-md-6">
                <Button bsStyle="default"
                      onClick={this.handleIgnoreChanges}>Ignore changes</Button>
              </div>
            </div>
          </fieldset>
        )
      } else {
        return (
          <fieldset>
            <legend>User defined data</legend>
            <div className="form-group">
              <table>
                <tr><td className="datalabel">Comment</td><td className="datafield">{this.state.Comment}</td></tr>
              </table>
            </div>

            <div className="form-group">
              <div className="col-md-6">
                <Button bsStyle="primary"
                  onClick={this.handleEdit}>Edit</Button>
              </div>

              <div className="col-md-6">
                <Button bsStyle="link"
                  onClick={this.handleToggle}>
                  Forget this agent
                </Button>
              </div>
            </div>
          </fieldset>
        )
      }
    },

    handleEdit: function(e) {
      this.setState({isEdit: true});
    },

    handleIgnoreChanges: function(e) {
      this.setState({isEdit: false});
    },

    render:function(){
      return (
          <div>
              <h2>System Information</h2>
              <table>
                <tr><td className="datalabel">Identifier</td><td className="datafield">{this.state.UUID}</td></tr>
                <tr><td className="datalabel">MachineGUID</td><td className="datafield">{this.state.MachineGUID}</td></tr>
                <tr><td className="datalabel">Agent Version</td><td className="datafield">{this.state.AgentVersion}</td></tr>
                <tr><td className="datalabel">OS</td><td className="datafield">{this.state.OS}</td></tr>
                <tr><td className="datalabel">Manufacturer</td><td className="datafield">{this.state.Manufacturer}</td></tr>
                <tr><td className="datalabel">Model</td><td className="datafield">{this.state.Model}</td></tr>
                <tr><td className="datalabel">First Seen</td><td className="datafield">{this.state.FirstSeen}</td></tr>
                <tr><td className="datalabel">Last Seen</td><td className="datafield">{this.state.LastSeen}</td></tr>
              </table>

              <Flash ref="flash"/>

              <form>
                  {this.showUserData()}
              </form>
          </div>
        )
    },


    // This is called by the `OverlayMixin` when this component
    // is mounted or updated and the return value is appended to the body.
    renderOverlay: function () {
      if (!this.state.isModalOpen) {
        return <span/>;
      }

      return (
          <Modal bsStyle="primary" title="Forget agent?" onRequestHide={this.handleToggle}>
            <div className="modal-body">
              <p>Are you sure you want to forget this agent?  It will no longer be visible in any views.</p>

              <p>You should only forget agents for systems that no longer exist, for example an agent that was on a system that was wiped.</p>

              <p>Once you forget an agent this cannot be undone.</p>
            </div>
            <div className="modal-footer">
              <Button bsStyle="danger" onClick={this.handleForget}>Yes, forget this agent</Button>
              <Button onClick={this.handleToggle}>No, don't forget this agent</Button>
            </div>
          </Modal>
        );
    }
  });
module.exports = SystemInfo;
