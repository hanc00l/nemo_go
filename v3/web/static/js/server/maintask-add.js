$(function () {
    load_select_list('/maintask-profile-list', $('#profile_name'));
    load_select_list('/org-all-list', $('#org_id'));
    init_profile_info();

    $("#profile_name").change(function () {
        // 获取当前选中的值
        const selectedValue = $(this).val();
        if (selectedValue === "") {
            init_profile_info();
        } else {
            show_profile_info(selectedValue);
        }
        $('#profile-tab').tab('show');
    });
    $('#target_split').change(function () {
        if ($(this).val() !== "0") {
            $('#target_split_num').prop('disabled', false);
        } else {
            $('#target_split_num').prop('disabled', true);
        }
    });
    $('#editProfile').click(function () {
        const selectedValue = $('#profile_name').val();
        if (selectedValue === "") {
            alert("请先选择配置模板！");
            return;
        }
        // 打开新页面
        window.open('/profile-edit?btn=close&id=' + selectedValue, '_blank');
    });
    $('#generateMaintaskName').click(function () {
        generate_maintask_name();
    })
    $('#maintaskForm').on('submit', function (e) {
        e.preventDefault();
        save_maintask_data();
    });
});

function process_form_data() {
    const target = $('#target').val();
    if (!isNotEmpty(target)) {
        alert("目标不能为空！");
        return null;
    }
    // 获取选中的 profile_name 和 profile_id
    const profileSelect = $('#profile_name');
    const profileId = profileSelect.val(); // 获取选中的 value（profile_id）
    const profileName = profileSelect.find('option:selected').text(); // 获取选中的文本（profile_name）
    const mainTaskInfoData = {
        //id: $('#id').val(), // 如果有 id 字段的话
        name: $('#name').val(),
        description: $('#description').val(),
        profile_name: profileName,
        profile_id: profileId,
        target: target.replace(/\r/g, "").replace(/\n/g, ","),
        exclude_target: $('#exclude_target').val(),
        target_split: parseInt($('#target_split').val(), 10),
        target_split_num: parseInt($('#target_split_num').val(), 10),
        org_id: $('#org_id').val(),
        is_cron_task: $('#is_cron_task').is(':checked'),
        cron_expr: $('#cron_exp').val(),
        is_proxy: $('#is_proxy').is(':checked'),
    };
    if (!isNotEmpty(mainTaskInfoData.name)) {
        alert("任务名称不能为空！");
        return null;
    }
    if (!isNotEmpty(mainTaskInfoData.profile_id)) {
        alert("请选择配置模板！");
        return null;
    }
    if (!isNotEmpty(mainTaskInfoData.target)) {
        alert("目标不能为空！");
        return null;
    }
    if (mainTaskInfoData.target_split !==0 && mainTaskInfoData.target_split_num < 1) {
        alert("拆分数量必须大于等于1！");
        return null;
    }
    if (mainTaskInfoData.is_cron_task && !isNotEmpty(mainTaskInfoData.cron_expr)) {
        alert("定时任务表达式不能为空！");
        return null;
    }
    return mainTaskInfoData;
}

function fill_form_with_data(data) {
    if (data.config_type === "staged") {
        $('#div_standalone').hide();
        // llmapi部分
        if (data.llmapi.enabled) {
            $('#div_llmapi').show();
            $('#qwen').prop('checked', data.llmapi.qwen);
            $('#kimi').prop('checked', data.llmapi.kimi);
            $('#deepseek').prop('checked', data.llmapi.deepseek);
            $('#icpPlus').prop('checked', data.llmapi.icpPlus);
            $('#llmapi_autoAssociateOrg').prop('checked', data.llmapi.config?.autoAssociateOrg || false);
        } else {
            $('#div_llmapi').hide();
        }
        // 端口扫描部分
        if (data.portscan.enabled) {
            $('#div_portscan').show();
            $('#nmap').prop('checked', data.portscan.nmap);
            $('#masscan').prop('checked', data.portscan.masscan);
            $('#gogo').prop('checked', data.portscan.gogo);
            $('#portscan_port').val(data.portscan.config.port);
            $('#portscan_rate').val(data.portscan.config.rate);
            $('#portscan_tech').val(data.portscan.config.tech);
            $('#portscan_PortscanConfig_maxOpenedPortPerIp').val(data.portscan.config.maxOpenedPortPerIp);
            $('#portscan_is_ping').prop('checked', data.portscan.config?.ping || false);
        } else {
            $('#div_portscan').hide();
        }

        // 域名扫描部分
        if (data.domainscan.enabled) {
            $('#div_domainscan').show();
            $('#massdns').prop('checked', data.domainscan.massdns);
            $('#subfinder').prop('checked', data.domainscan.subfinder);
            $('#domainscan_is_ignore_cdn').prop('checked', data.domainscan.config?.ignorecdn || false);
            $('#domainscan_is_ignore_china_other').prop('checked', data.domainscan.config?.ignorechinaother || false);
            $('#domainscan_is_ignore_outside_china').prop('checked', data.domainscan.config?.ignoreoutsidechina || false);
            $('#domainscan_is_ip_port_scan').prop('checked', data.domainscan.config?.portscan || false);
            $('#domainscan_is_ip_subnet_port_scan').prop('checked', data.domainscan.config?.subnetPortscan || false);
            $('#domainscan_PortscanConfig_maxResolvedDomainPerIP').val(data.domainscan.config.maxResolvedDomainPerIP);
            $('#domainscan_wordlist_file').val(data.domainscan.config.wordlistFile);
            if (data.domainscan.config.portscan || data.domainscan.config.subnetPortscan) {
                $('#domain_portscan_config').show();
                $('#domain_portscan_bin').val(data.domainscan.config.resultPortscanBin);
                $('#domain_portscan_port').val(data.domainscan.config.resultPortscanConfig.port);
                $('#domain_portscan_rate').val(data.domainscan.config.resultPortscanConfig.rate);
                $('#domain_portscan_tech').val(data.domainscan.config.resultPortscanConfig.tech);
                $('#domain_portscan_PortscanConfig_maxOpenedPortPerIp').val(data.domainscan.config.resultPortscanConfig.maxOpenedPortPerIp);
            } else {
                $('#domain_portscan_config').hide();
            }
        } else {
            $('#div_domainscan').hide();
        }
        // 在线API部分
        if (data.onlineapi.enabled) {
            $('#div_onlineapi').show();
            $('#fofa').prop('checked', data.onlineapi.fofa);
            $('#hunter').prop('checked', data.onlineapi.hunter);
            $('#quake').prop('checked', data.onlineapi.quake);
            $('#whois').prop('checked', data.onlineapi.whois);
            $('#icp').prop('checked', data.onlineapi.icp);
            $('#onlineapi_search_by_keyword').prop('checked', data.onlineapi.config?.searchbykeyword || false);
            $('#onlineapi_is_ignore_cdn').prop('checked', data?.onlineapi.config?.ignorecdn || false);
            $('#onlineapi_is_ignore_china_other').prop('checked', data.onlineapi.config?.ignorechinaother || false);
            $('#onlineapi_is_ignore_outside_china').prop('checked', data.onlineapi.config?.ignoreoutsidechina || false);
            $('#onlineapi_search_start_time').val(data.onlineapi.config.searchstarttime);
            $('#onlineapi_search_limit_count').val(data.onlineapi.config.searchlimitcount);
            $('#onlineapi_search_page_size').val(data.onlineapi.config.searchpagesize);
        } else {
            $('#div_onlineapi').hide();
        }
        // 指纹识别部分
        if (data.fingerprint.enabled) {
            $('#div_fingerprint').show();
            $('#fingerprint_httpx').prop('checked', data.fingerprint.config?.httpx || false);
            $('#fingerprint_screenshot').prop('checked', data.fingerprint.config?.screenshot || false);
            $('#fingerprint_icon_hash').prop('checked', data.fingerprint.config?.iconhash || false);
            $('#fingerprint_fingerprintx').prop('checked', data.fingerprint.config?.fingerprintx || false);
        } else {
            $('#div_fingerprint').hide();
        }
        // POC扫描部分
        if (data.pocscan.enabled) {
            $('#div_pocscan').show();
            $('#pocscan_bin').val(data.pocscan.pocbin);
            $('#pocscan_poc_type').val(data.pocscan.config.pocType);
            $('#pocscan_base_web_status').prop('checked', data.pocscan.config.baseWebStatus);
            $('#pocscan_poc_file').val(data.pocscan.config.pocFile);
            $('#pocscan_brute_password').prop('checked', data.pocscan.config.brutePassword);
        } else {
            $('#div_pocscan').hide();
        }
    } else if (data.config_type === "standalone") {
        $('#div_standalone').show();
        $('#div_portscan').hide();
        $('#div_domainscan').hide();
        $('#div_onlineapi').hide();
        $('#div_fingerprint').hide();
        $('#div_pocscan').hide();
        // 独立扫描部分
        $('#standalone_port').val(data.standalone.config.port);
        $('#standalone_rate').val(data.standalone.config.rate);
        $('#standalone_PortscanConfig_maxOpenedPortPerIp').val(data.standalone.config.maxOpenedPortPerIp);
        $('#standalone_is_ping').prop('checked', data.standalone.config?.ping || false);
        $('#standalone_is_verbose').prop('checked', data.standalone.config?.verbose || false);
        $('#standalone_is_poc_scan').prop('checked', data.standalone.config?.pocscan || false);
    }
}

function save_maintask_data() {
    const maintaskInfoData = process_form_data();
    if (maintaskInfoData === null) {
        return;
    }
    $.ajax({
        type: 'POST',
        url: '/maintask-save',
        contentType: 'application/json',
        data: JSON.stringify(maintaskInfoData),
        success: function (data) {
            show_response_message("保存成功", data, function () {
                history.back();
            })
        },
        error: function (xhr, status, error) {
            swal('失败', '请求失败: ' + error, 'error');
        }
    });
}

function show_profile_info(id) {
    $.ajax({
        type: 'POST',
        url: '/profile-get', // Beego处理URL
        data: {id: id},
        dataType: 'json',
        success: function (response) {
            fill_form_with_data(response);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function init_profile_info() {
    // 设置任务模板的默认配置
    $('#div_llmapi').hide();
    $('#div_portscan').hide();
    $('#div_domainscan').hide();
    $('#div_onlineapi').hide();
    $('#div_fingerprint').hide();
    $('#div_pocscan').hide();
    $('#div_standalone').hide();
    $('#llmapi_config input, #llmapi_config select').prop('disabled', true);
    $('#portscan_config input, #portscan_config select').prop('disabled', true);
    $('#domainscan_config input, #domainscan_config select').prop('disabled', true);
    $('#onlineapi_config input, #onlineapi_config select').prop('disabled', true);
    $('#fingerprint_config input, #fingerprint_config select').prop('disabled', true);
    $('#pocscan_config input, #pocscan_config select').prop('disabled', true);
    $('#standalone_config input, #standalone_config select').prop('disabled', true);
}

function generate_maintask_name() {
    const profileName = $('#profile_name').find('option:selected').text();
    const targetSplit = $('#target_split').val();
    const isCronTask = $('#is_cron_task').is(':checked');
    const isProxy = $('#is_proxy').is(':checked');

    let name = profileName;
    if (targetSplit === "2") {
        name += "-按IP分片"
    }
    if (isCronTask) {
        name = "[定时]-" + name;
    }
    if (isProxy) {
        name += "-[代理]"
    }

    $('#name').val(name);
}
