var React = require('react');
var DashboardItem = require('./dashboard-item.jsx');

var Dashboard =
  React.createClass({
    render:function(){
      return (
          <div>
            <div className="row">
              <div className="col-md-12">
                <div className="panel panel-default">
                  <div className="panel-body">
                    <h2>Summit Route &mdash; EPP</h2>
                    <p>
                      <a href="/download/SREPP.exe"><span className="glyphicon glyphicon-download"></span> Download your agent</a>
                      &mdash; Your agent installer is pre-configured so the data it collects will show up in your dashboard.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )
    }
  });
module.exports = Dashboard;
