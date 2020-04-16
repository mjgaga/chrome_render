chrome.browserAction.onClicked.addListener(function () {
  // chrome.tabs.getSelected(null, function (tab) {
      captureTab();
  // });
});
// alert(chrome.tabs)
// chrome.tabs.onCreated.addListener(function (tab) {
//   (function fn(tab) {
//     alert(tab.status)
//     if (tab.status == "complete") {
setTimeout(captureTab, 1000)
//     } else {
//       setTimeout(fn.bind(null, tab), 1000)
//     }
//   })(tab);
// });


const ws = new WebSocket("ws://127.0.0.1:8080")
ws.binaryType = 'arraybuffer';

var recordedChunks = [];

function captureTab() {
    const constraints = {
        audio: true,
        video: true,
        audioConstraints: {
            mandatory: {
                chromeMediaSource: 'tab',
                echoCancellation: true
            }
        },
        videoConstraints: {
            mandatory: {
                chromeMediaSource: 'tab',
                maxWidth: 1280,
                minWidth: 1280,
                maxFrameRate: 15,
                minAspectRatio: 1.77,

            }
        }
    };

    chrome.tabCapture.capture(constraints, function (stream) {
        if (!stream) {
            console.error("couldn't record tab");
            return;
        }
        const originalAudioTrack = stream.getAudioTracks()[0];
        // const originalVideoTrack = stream.getVideoTracks()[0];

        const audioCtx = new AudioContext();
        const source = audioCtx.createMediaStreamSource(stream);
        const gainFilter = audioCtx.createGain()
        const destination = audioCtx.createMediaStreamDestination()
        const outputStream = destination.stream

        source.connect(gainFilter);
        gainFilter.connect(destination)

        const filteredTrack = outputStream.getAudioTracks()[0]
        stream.addTrack(filteredTrack)
        stream.removeTrack(originalAudioTrack)


        const recorder = new MediaRecorder(stream, {
            mimeType: 'video/webm',
            videoBitsPerSecond:  1000 * 1000,
            audioBitsPerSecond:  64 * 1000
        });

        recorder.start(1000);
        recorder.ondataavailable = handleDataAvailable

        // setTimeout(event => {
        //     console.log("stopping");
        //     recorder.stop();
        // }, 10000);
    });
}

function handleDataAvailable(event) {
  console.log("data-available");

  if (event.data.size > 0) {
      blobToArrayBufferConverter([event.data], (arrBuffer) => {
          ws.send(arrBuffer);
      })
    // recordedChunks.push(event.data);
    // console.log(recordedChunks);
    // download();
  } else {

  }
}

function blobToArrayBufferConverter(blobChunks,callback){
    var blob = new Blob(blobChunks, {type: 'video/webm'});
    var fileReader = new FileReader();
    fileReader.readAsArrayBuffer(blob);
    fileReader.onload  = function(progressEvent) {
        callback(this.result);
    }
}

// var tname = Date.now();
// function download() {
//   var blob = new Blob(recordedChunks, {
//     type: "video/webm"
//   });
//   var url = URL.createObjectURL(blob);
//   var a = document.createElement("a");
//   document.body.appendChild(a);
//   a.style = "display: none";
//   a.href = url;
//   a.download = "abc_" + tname + ".webm";
//   a.click();
//   window.URL.revokeObjectURL(url);
// }

