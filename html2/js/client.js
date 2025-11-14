var cut = document.querySelector("#cut")
var canvas = document.querySelector("#canvas")
var video = document.querySelector("#video")

//获取设备
if (navigator.mediaDevices && navigator.mediaDevices.enumerateDevices()) {
    navigator.mediaDevices.enumerateDevices()
        .then(deviceInfos)
        .catch(handleError)
}

function deviceInfos(devices) {
    console.log(devices)
    devices.forEach(function (device) {
        console.log(device.kind, device.label, device.groupId, device.deviceId)
    })
}

function handleError(err) {
    console.log(err)
}

//获取音视频
if (navigator.mediaDevices.getUserMedia) {
    navigator.mediaDevices.getUserMedia({
        video: true,
        audio: true
    })
        .then(handleMediaStream)
        .catch(handleError)
}

function handleMediaStream(stream) {
    video.srcObject = stream
}

cut.onclick = function () {
    canvas.getContext('2d').drawImage(video, 0, 0, 320, 240)
}

let rtcPeerConnection = new RTCPeerConnection();
rtcPeerConnection.createDataChannel()
