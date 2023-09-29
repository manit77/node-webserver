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

  //start the http server
  let port = 80;
  httpServer.listen(port, () => {
    console.log(`http://localhost:${port}`);
  });
  
})();


