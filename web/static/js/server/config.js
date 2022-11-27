$(function () {
    $('#btnsiderbar').click();
    load_config();
    load_custom('honeypot', $('#text_honeypot'));
    load_custom('service', $('#text_service'));
    load_custom('iplocation', $('#text_iplocation'));
    load_custom('iplocationB', $('#text_iplocationB'));
    load_custom('iplocationC', $('#text_iplocationC'));

    $("#buttonSaveNmap").click(function () {
        $.post("/config-save-portscan",
            {
                "cmdbin": $('#select_cmdbin').val(),
                "port": $('#input_port').val(),
                "rate": $('#input_rate').val(),
                "tech": $('#select_tech').val(),
                "ping": $('#checkbox_ping').is(":checked"),
            }, function (data, e) {
                if (e === "success" && data['status'] == 'success') {
                    swal({
                        title: "保存成功！",
                        text: "",
                        type: "success",
                        confirmButtonText: "确定",
                        confirmButtonColor: "#41b883",
                        closeOnConfirm: true,
                        timer: 3000
                    });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });
    });

    $("#buttonSaveFingerprint").click(function () {
        $.post("/config-save-fingerprint",
            {
                "httpx": $('#checkbox_httpx').is(":checked"),
                "fingerprinthub": $('#checkbox_fingerprinthub').is(":checked"),
                "screenshot": $('#checkbox_screenshot').is(":checked"),
                "iconhash": $('#checkbox_iconhash').is(":checked"),

            }, function (data, e) {
                if (e === "success" && data['status'] == 'success') {
                    swal({
                        title: "保存成功！",
                        text: "",
                        type: "success",
                        confirmButtonText: "确定",
                        confirmButtonColor: "#41b883",
                        closeOnConfirm: true,
                        timer: 3000
                    });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });
    });

    $("#buttonSaveTaskSlice").click(function () {
        if ($('#input_ipslicenumber').val() === '' || $('#input_portslicenumber').val() === '') {
            swal('Warning', "请输入数量", 'error');
            return;
        }
        $.post("/config-save-taskslice",
            {
                "portslicenumber": $('#input_portslicenumber').val(),
                "ipslicenumber": $('#input_ipslicenumber').val(),
            }, function (data, e) {
                if (e === "success" && data['status'] == 'success') {
                    swal({
                        title: "保存成功！",
                        text: "",
                        type: "success",
                        confirmButtonText: "确定",
                        confirmButtonColor: "#41b883",
                        closeOnConfirm: true,
                        timer: 3000
                    });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });
    });

    $("#buttonSaveNotify").click(function () {
        $.post("/config-save-notify",
            {
                "token_serverchan": $('#input_serverchan').val(),
                "token_dingtalk": $('#input_dingtalk').val(),
                "token_feishu": $('#input_feishu').val(),
            }, function (data, e) {
                if (e === "success" && data['status'] == 'success') {
                    swal({
                        title: "保存成功！",
                        text: "",
                        type: "success",
                        confirmButtonText: "确定",
                        confirmButtonColor: "#41b883",
                        closeOnConfirm: true,
                        timer: 3000
                    });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });
    });
});
$("#buttonSaveHoneypot").click(function () {
    save_custom("honeypot", $('#text_honeypot').val())
});
$("#buttonSaveService").click(function () {
    save_custom("service", $('#text_service').val())
});
$("#buttonSaveIPLocation").click(function () {
    save_custom("iplocation", $('#text_iplocation').val())
});
$("#buttonSaveIPLocationB").click(function () {
    save_custom("iplocationB", $('#text_iplocationB').val())
});
$("#buttonSaveIPLocationC").click(function () {
    save_custom("iplocationC", $('#text_iplocationC').val())
});
$("#buttonUploadXrayPoc").click(function () {
    let formData = new FormData();
    formData.append('file', $('#file_xraypoc')[0].files[0]);
    if (formData.get("file") === "undefined") {
        swal('Warning', '请选择要上传的文件！', 'error');
        return;
    }
    $.ajax({
        url: "/config-upload-xraypoc",
        type: "post",
        data: formData,
        //十分重要，不能省略
        cache: false,
        processData: false,
        contentType: false,
        success: function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                swal({
                    title: "保存成功！",
                    text: data['msg'],
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                    timer: 3000
                });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        }
    });
});
$("#buttonChangPassword").click(function () {
    if ($('#input_oldpass').val() === '' || $('#input_password1').val() === '' || $('#input_password2').val() === '') {
        swal('Warning', "请输入密码！", 'error');
        return;
    }
    if ($('#input_password1').val() !== $('#input_password2').val()) {
        swal('Warning', "两次新密码不一致！", 'error');
        return;
    }
    $.post("/config-change-password",
        {
            "oldpass": $('#input_oldpass').val(),
            "newpass": $('#input_password1').val()
        }, function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                $('#input_oldpass').val('');
                $('#input_password1').val('');
                $('#input_password2').val('');
                swal({
                    title: "密码修改成功！",
                    text: "",
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                    timer: 3000
                });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        });
});

function load_config() {
    $.post("/config-list", function (data) {
        $('#select_cmdbin').val(data['cmdbin']);
        $('#input_port').val(data['port']);
        $('#select_tech').val(data['tech']);
        $('#input_rate').val(data['rate']);
        $('#checkbox_ping').prop("checked", data['ping']);
        $('#checkbox_httpx').prop("checked", data['httpx']);
        $('#checkbox_fingerprinthub').prop("checked", data['fingerprinthub']);
        $('#checkbox_screenshot').prop("checked", data['screenshot']);
        $('#checkbox_iconhash').prop("checked", data['iconhash']);
        $('#input_ipslicenumber').val(data['ipslicenumber']);
        $('#input_portslicenumber').val(data['portslicenumber']);
        $('#nemo_version').html(data['version']);
        $('#input_serverchan').val(data['serverchan']);
        $('#input_dingtalk').val(data['dingtalk']);
        $('#input_feishu').val(data['feishu']);
    });
}

function load_custom(type, textCtl) {
    $.post("/custom-load",
        {
            "type": type,
        }, function (data, e) {
            if (e === "success" && data['status'] === "success") {
                textCtl.val(data['msg']);
            } else {
                textCtl.val('load custom fail!');
            }
        });
}

function save_custom(type, content) {
    if (content === "") {
        swal('Warning', "空的配置！", 'error');
        return;
    }
    $.post("/custom-save",
        {
            "type": type,
            "content": content
        }, function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                swal({
                    title: "保存成功！",
                    text: "",
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                    timer: 3000
                });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        });
}