// for mongo v4
rs.initiate( {
   _id : "myReplicaSet",
   members: [
     { _id: 0, host: "mongo1:27017", priority: 2 },
     { _id: 1, host: "mongo2:27017", priority: 1 },
     { _id: 2, host: "mongo3:27017", priority: 1 }
   ]
})


// // for mongo v5
// rs.initiate({
//   _id: "myReplicaSet",
//   members: [
//     { _id: 0, host: "mongo1:27017", votes: 1 },
//     { _id: 1, host: "mongo2:27017" },
//     { _id: 2, host: "mongo3:27017" }
//   ]
// })


