var React = require('react');
var Reqwest = require("reqwest");
var qs = require('querystring');

var File =
  React.createClass({
    getInitialState: function() {
      return {
        sha256: '', sha1: '', md5: '',
        size: '',
        filepath: '',
        firstseen: '',
        lastseen: '',
        numsystems: '',

        companyname: '',
        productversion: '',
        productname: '',
        filedescription: '',
        internalname: '',
        fileversion: '',
        originalfilename: '',

        subjectshortname: '',
        subject: '',
        serialnumber: '',
        digestalgorithm: '',
        digestencryptionalgorithm: ''
        };
    },

    componentDidMount:function(){
      var thisComponent = this;
      var query = qs.parse(window.location.search.slice(1));

      Reqwest({
        url: '/api/fileinfo.json?sha256='+query["sha256"],
        type: 'json',
        success: function (resp) {
          if (thisComponent.isMounted()) {
            thisComponent.setState({
             sha256: resp.Sha256,
             sha1: resp.Sha1,
             md5: resp.Md5,

             filepath: resp.FilePath,
             firstseen: resp.FirstSeen,
             lastseen: resp.LastSeen,
             numsystems: resp.NumSystems,

             size: resp.Size,
             companyname: resp.CompanyName,
             productversion: resp.ProductVersion,
             productname: resp.ProductName,
             filedescription: resp.FileDescription,
             internalname: resp.InternalName,
             fileversion: resp.FileVersion,
             originalfilename: resp.OriginalFilename,

             subjectshortname: resp.SubjectShortName,
             subject: resp.Subject,
             serialnumber: resp.SerialNumber,
             digestalgorithm: resp.DigestAlgorithm,
             digestencryptionalgorithm: resp.DigestEncryptionAlgorithm
            });
          }
        }
      });
    },

    render:function(){
      return (
          <div>
              <h2>File</h2>

              <h3>Your network</h3>
              <table className="data_listing">
                <tr><th>File path</th><td>{this.state.filepath}</td></tr>
                <tr><th>First seen</th><td>{this.state.firstseen}</td></tr>
                <tr><th>Last seen</th><td>{this.state.lastseen}</td></tr>
                <tr><th>Number of systems</th><td>{this.state.numsystems}</td></tr>
              </table>

              <h3>File Data</h3>
              <div className="row">
                <div className="col-md-7">
                <h4>Calculated from the file</h4>
                  <table className="data_listing">
                    <tr><th>Sha256</th><td>{this.state.sha256}</td></tr>
                    <tr><th>Sha1</th><td>{this.state.sha1}</td></tr>
                    <tr><th>Md5</th><td>{this.state.md5}</td></tr>
                    <tr><th>Size</th><td>{this.state.size}</td></tr>
                  </table>
                </div>

                <div className="col-md-5">
                  <h4>Extracted from the file</h4>
                  <table className="data_listing">
                    <tr><th>Company Name</th><td>{this.state.companyname}</td></tr>
                    <tr><th>Product Name</th><td>{this.state.productname}</td></tr>
                    <tr><th>Product Version</th><td>{this.state.productversion}</td></tr>
                    <tr><th>File Description</th><td>{this.state.filedescription}</td></tr>
                    <tr><th>Internal name</th><td>{this.state.internalname}</td></tr>
                    <tr><th>File version</th><td>{this.state.fileversion}</td></tr>
                    <tr><th>Original filename</th><td>{this.state.originalfilename}</td></tr>
                  </table>
                </div>
              </div>

              <h3>Signature information</h3>
              <table className="data_listing">
                <tr><th>Subject</th><td>{this.state.subject}</td></tr>
                <tr><th>Serial Number</th><td>{this.state.serialnumber}</td></tr>
                <tr><th>Digest algorithm</th><td>{this.state.digestalgorithm}</td></tr>
                <tr><th>Digest encryption algorithm</th><td>{this.state.digestencryptionalgorithm}</td></tr>
              </table>

          </div>
        )
    }
  });
module.exports = File;
