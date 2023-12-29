$(function () {
    //$('#btnsiderbar').click();
    load_config();
    load_custom('task_workspace', $('#text_task_workspace'));
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
                "fingerprintx": $('#checkbox_fingerprintx').is(":checked"),
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
    $("#buttonTestNotify").click(function () {
        $.post("/config-test-notify", {}, function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                swal({
                    title: "测试完成",
                    text: data['msg'],
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        });
    });
    $("#buttonSaveAPIToken").click(function () {
        $.post("/config-save-api",
            {
                "fofa": $('#checkbox_fofa').is(":checked"),
                "hunter": $('#checkbox_hunter').is(":checked"),
                "quake": $('#checkbox_quake').is(":checked"),
                "fofatoken": $('#input_fofa_token').val(),
                "huntertoken": $('#input_hunter_token').val(),
                "quaketoken": $('#input_quake_token').val(),
                "chinaztoken": $('#input_chinaz_token').val(),
                "pagesize": $('#input_pagesize').val(),
                "limitcount": $('#input_limitcount').val(),
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
    $("#buttonTestAPIToken").click(function () {
        $.post("/config-test-api", {}, function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                swal({
                    title: "测试完成",
                    text: data['msg'],
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        });
    });
    $("#buttonSaveDomainscan").click(function () {
        $.post("/config-save-domainscan",
            {
                "subfinder": $('#checkbox_subfinder').is(":checked"),
                "subdomainbrute": $('#checkbox_subdomainbrute').is(":checked"),
                "subdomaincrawler": $('#checkbox_subdomaincrawler').is(":checked"),
                "icp": $('#checkbox_icp').is(":checked"),
                "whois": $('#checkbox_whois').is(":checked"),
                "portscan": $('#checkbox_portscan').is(":checked"),
                "ignorecdn": $('#checkbox_ignorecdn').is(":checked"),
                "ignoreoutofchina": $('#checkbox_ignoreoutofchina').is(":checked"),
                "wordlist": $('#select_wordlist').val(),
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
    $("#buttonSaveWorkerProxy").click(function () {
        $.post("/config-save-workerproxy",
            {
                "proxyList": $('#text_proxy_list').val(),
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
    $("#buttonChangPassword").click(function () {
        if ($('#input_oldpass').val() === '' || $('#input_password1').val() === '' || $('#input_password2').val() === '') {
            swal('Warning', "请输入密码！", 'error');
            return;
        }
        if ($('#input_password1').val() !== $('#input_password2').val()) {
            swal('Warning', "两次新密码不一致！", 'error');
            return;
        }
        let encryptor = new JSEncrypt();
        encryptor.setPublicKey(pubKey);
        let oldpass = encryptor.encrypt($('#input_oldpass').val());
        let newpass = encryptor.encrypt($('#input_password1').val());
        $.post("/config-change-password",
            {
                "oldpass": oldpass,
                "newpass": newpass,
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
    $("#buttonSaveTaskWorkspace").click(function () {
        save_custom("task_workspace", $('#text_task_workspace').val(), "/custom-save-taskworkspace")
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
        $('#checkbox_fingerprintx').prop("checked", data['fingerprintx']);

        $('#input_ipslicenumber').val(data['ipslicenumber']);
        $('#input_portslicenumber').val(data['portslicenumber']);
        $('#nemo_version').html(data['version']);

        $('#input_serverchan').val(data['serverchan']);
        $('#input_dingtalk').val(data['dingtalk']);
        $('#input_feishu').val(data['feishu']);
        $('#input_fofa_token').val(data['fofatoken']);
        $('#input_hunter_token').val(data['huntertoken']);
        $('#input_quake_token').val(data['quaketoken']);
        $('#input_chinaz_token').val(data['chinaztoken']);

        $('#checkbox_subfinder').prop("checked", data['subfinder']);
        $('#checkbox_subdomainbrute').prop("checked", data['subdomainbrute']);
        $('#checkbox_subdomaincrawler').prop("checked", data['subdomaincrawler']);
        $('#checkbox_icp').prop("checked", data['icp']);
        $('#checkbox_whois').prop("checked", data['whois']);
        $('#checkbox_ignorecdn').prop("checked", data['ignorecdn']);
        $('#checkbox_ignoreoutofchina').prop("checked", data['ignoreoutofchina']);
        $('#checkbox_portscan').prop("checked", data['portscan']);
        $('#select_wordlist').val(data['wordlist']);

        $('#checkbox_fofa').prop("checked", data['fofa']);
        $('#checkbox_hunter').prop("checked", data['hunter']);
        $('#checkbox_quake').prop("checked", data['quake']);
        $('#input_pagesize').val(data['pagesize']);
        $('#input_limitcount').val(data['limitcount']);

        $("#text_proxy_list").val(data['proxyList']);
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

function save_custom(type, content, url = "/custom-save") {
    if (type === "") {
        swal('Warning', "类型为空！", 'error');
        return;
    }
    $.post(url,
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
