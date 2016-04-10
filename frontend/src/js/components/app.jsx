var React = require('react');
var Router = require('react-router-component');

var Template = require('./app-template.jsx');
var Dashboard = require('./dashboard/app-dashboard.jsx');

var Systems = require('./systems/app-systems.jsx');
var SystemInfo = require('./systems/app-systeminfo.jsx');

var ProcessEvents = require('./process_events/app-process_events.jsx');

var File = require('./file/app-file.jsx');
var Executables = require('./executables/app-executables.jsx');

var Profile = require('./profile/app-profile.jsx');
var PasswordChange = require('./password_change/app-password_change.jsx');
var PasswordReset = require('./password_reset/app-password_reset.jsx');

var PrivacyPolicy = require('./app-privacy_policy.jsx');
var TermsAndConditions = require('./app-terms_and_conditions.jsx');
var Help = require('./app-help.jsx');

var Locations = Router.Locations;
var Location = Router.Location;
var NotFound = Router.NotFound;

var APP =
  React.createClass({
    render:function(){
      return (
        <Template>
          <Locations>
            <Location path="/" handler={Dashboard} />
            <Location path="/systems" handler={Systems} />
            <Location path="/systeminfo*" handler={SystemInfo} />
            <Location path="/process_events" handler={ProcessEvents} />
            <Location path="/file*" handler={File} />
            <Location path="/executables" handler={Executables} />
            <Location path="/profile" handler={Profile} />
            <Location path="/change_password" handler={PasswordChange} />
            <Location path="/password_reset" handler={PasswordReset} />
            <Location path="/privacy_policy" handler={PrivacyPolicy} />
            <Location path="/terms_and_conditions" handler={TermsAndConditions} />
            <Location path="/help" handler={Help} />
            <NotFound handler={Dashboard} />
          </Locations>
        </Template>
      )
    }
  });
module.exports = APP;
