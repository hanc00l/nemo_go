$(function () {
    //$('#btnsiderbar').click();
    load_config_worker();
    $("#buttonSaveNmap").click(function () {
        $.post("/config-save-portscan",
            {
                "cmdbin": $('#select_cmdbin').val(),
                "port": $('#input_port').val(),
                "rate": $('#input_rate').val(),
                "tech": $('#select_tech').val(),
                "ping": $('#checkbox_ping').is(":checked"),
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
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
                if (e === "success" && data['status'] === 'success') {
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
                if (e === "success" && data['status'] === 'success') {
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
            if (e === "success" && data['status'] === 'success') {
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
                if (e === "success" && data['status'] === 'success') {
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
                if (e === "success" && data['status'] === 'success') {
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
    $("#buttonSaveTaskFilter").click(function () {
        $.post("/config-save-workerfilter",
            {
                "maxportperip": $('#input_maxportperip').val(),
                "maxdomainperip": $('#input_maxdomainperip').val(),
                "title": $('#input_title').val(),
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
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

function load_config_worker() {
    $.post("/config-list-worker", function (data) {
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

        $('#input_maxportperip').val(data['maxportperip']);
        $('#input_maxdomainperip').val(data['maxdomainperip']);
        $('#input_title').val(data['title']);

        $("#text_proxy_list").val(data['proxyList']);
    });
}