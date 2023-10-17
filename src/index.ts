import express from "express";
import cors from "cors";
const path = require('path');

(async () => {

  const expressApp = express();  
  const httpServer = require("http").createServer(expressApp);  
  
  expressApp.use(cors());
  expressApp.use(express.json({ limit: '100mb' }));
  expressApp.use(express.urlencoded({ extended: true }));
  expressApp.options('*', cors()) // include before other routes  
  expressApp.use(express.static(path.join(__dirname, "../public")));

  expressApp.get("/", async function (req, res, next) {
    res.write("hello from node-web-server");
    res.end();
  });

  //start the http server
  let port = "80";
  if(process.env.WEBSERVERPORT){
    port = process.env.WEBSERVERPORT;
  }
  httpServer.listen(port, () => {
    console.log(`http://localhost:${port}`);
  });
  
})();


