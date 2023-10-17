import express from "express";
import cors from "cors";
import * as http from "http"
import * as path from "path"

(async () => {

  const expressApp = express();  
  const httpServer = http.createServer(expressApp);  
  
  expressApp.use(cors());
  expressApp.use(express.json({ limit: '100mb' }));
  expressApp.use(express.urlencoded({ extended: true }));
  expressApp.options('*', cors()) // include before other routes  
  expressApp.use(express.static(path.join(__dirname, "../public")));

  //test hello
  expressApp.get("/", async function (req, res, next) {
    res.write(`hello from node-web-server ${new Date()}`);
    res.end();
  });

  //test json get
  expressApp.get("/json", async function (req, res, next) {
    res.json({
      message: `hello from node-web-server ${new Date()}`,
    });
  });

  //test json post
  expressApp.post("/json", async function (req, res, next) {
    res.json({
      message: `hello from node-web-server ${new Date()}`,
      post: req.body
    });
  });

  //start the http server
  let port = "80";
  if(process.env.WEB_SERVER_PORT){
    port = process.env.WEB_SERVER_PORT;
  }
  
  httpServer.listen(port, () => {
    console.log(`http://localhost:${port}`);
  });
  
})();


//process shutdown signal, exit 0
process.on('SIGINT', function() {
    process.exit(0);
});
