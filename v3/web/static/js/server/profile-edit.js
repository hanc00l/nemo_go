$(function () {
    init_pocfile_multiselect();
    const id = $('#hidden_id').val();
    if (id !== "") {
        load_profile_data(id);
    } else {
        load_pocfile_data(null);
        $('#llmapi_config input, #llmapi_config select').prop('disabled', true);
        $('#portscan_config input, #portscan_config select').prop('disabled', true);
        $('#domainscan_config input, #domainscan_config select').prop('disabled', true);
        $('#onlineapi_config input, #onlineapi_config select').prop('disabled', true);
        $('#fingerprint_config input, #fingerprint_config select').prop('disabled', true);
        $('#pocscan_config input, #pocscan_config select').prop('disabled', true);
        $('#pocscan_poc_file').multiselect('disable');
    }

    $.urlParam = function (param) {
        const results = new RegExp(`[?&]${param}=([^&#]*)`).exec(window.location.href);
        return results ? results[1] : null;
    };


    $('#btn_cancel').click(function () {
        const btnValue = $.urlParam('btn'); // 获取 btn 参数的值
        if (btnValue === "close") {
            window.close();
        } else {
            history.back();
        }
    });


    $('#profileForm').on('submit', function (e) {
        e.preventDefault();
        save_profile_data();
    });

    // 切换Tab时切换单选按钮
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
        const target = $(e.target).attr("href");
        if (target === "#staged") {
            $('input[name="config_type"][value="staged"]').prop('checked', true);
        } else if (target === "#standalone") {
            $('input[name="config_type"][value="standalone"]').prop('checked', true);
        }
    });

    // 切换单选按钮时切换Tab
    $('input[name="config_type"]').on('change', function () {
        if ($(this).val() === "staged") {
            $('#staged-tab').tab('show');
        } else if ($(this).val() === "standalone") {
            $('#standalone-tab').tab('show');
        }
    });

    $('#enable_llmapi').change(function () {
        const isEnabled = $(this).prop('checked');
        $('#llmapi_config input, #llmapi_config select').prop('disabled', !isEnabled);
    });

    $('#enable_portscan').change(function () {
        const isEnabled = $(this).prop('checked');
        $('#portscan_config input, #portscan_config select').prop('disabled', !isEnabled);
    });

    // 域名扫描
    $('#enable_domainscan').change(function () {
        const isEnabled = $(this).prop('checked');
        $('#domainscan_config input, #domainscan_config select').prop('disabled', !isEnabled);
    });

    // 在线API
    $('#enable_onlineapi').change(function () {
        const isEnabled = $(this).prop('checked');
        $('#onlineapi_config input, #onlineapi_config select').prop('disabled', !isEnabled);
    });

    // 指纹识别
    $('#enable_fingerprint').change(function () {
        const isEnabled = $(this).prop('checked');
        $('#fingerprint_config input, #fingerprint_config select').prop('disabled', !isEnabled);
    });

    // POC扫描
    $('#enable_pocscan').change(function () {
        const isEnabled = $(this).prop('checked');
        $('#pocscan_config input, #pocscan_config select').prop('disabled', !isEnabled);
        if (isEnabled && $('#pocscan_poc_type').val() === 'selectedPocFile') {
            $('#pocscan_poc_file').multiselect('enable')
        } else {
            $('#pocscan_poc_file').multiselect('disable')
        }
    });
    // 生成
    $('#generateProfileName').on('click', function () {
        const name = generate_profile_name();
        $('#name').val(name);
    });
    $('#pocscan_poc_type').change(function () {
        const pocType = $(this).val();
        if (pocType === 'matchFinger') {
            $('#pocscan_poc_file').multiselect('disable');
        } else {
            $('#pocscan_poc_file').multiselect('enable');
        }
    });
});


function load_pocfile_data(selectedPocFiles) {
    $.ajax({
        url: '/vul-load-pocfile', // 替换为你的 API 地址
        method: 'POST',
        data: {poc_bin: $('#pocscan_bin').val()},
        dataType: 'json',
        success: function (data) {
            // 清空现有的选项
            $('#pocscan_poc_file').empty();
            // 遍历 JSON 数据，动态生成 <option> 元素
            $.each(data, function (index, item) {
                var option = $('<option></option>')
                    .attr('value', item.text)
                    .text(item.name);
                $('#pocscan_poc_file').append(option);
            });
            // 初始化或刷新 Multiselect
            if (selectedPocFiles) {
                $('#pocscan_poc_file').val(selectedPocFiles);
                sortSelectedOptionsToTop();
            } else {
                $('#pocscan_poc_file').multiselect('rebuild');
            }
        },
        error: function (xhr, status, error) {
            console.error('加载选项失败:', error);
        }
    });
}

// 将选中的选项移动到最顶部
function sortSelectedOptionsToTop() {
    const $select = $('#pocscan_poc_file');
    const selectedOptions = $select.find('option:selected').get(); // 获取选中的选项
    const unselectedOptions = $select.find('option:not(:selected)').get(); // 获取未选中的选项

    // 清空下拉框
    $select.empty();

    // 先添加选中的选项
    $.each(selectedOptions, function (index, option) {
        $select.append(option);
    });

    // 再添加未选中的选项
    $.each(unselectedOptions, function (index, option) {
        $select.append(option);
    });

    // 刷新 Multiselect
    $select.multiselect('rebuild');
}

function init_pocfile_multiselect() {
    $('#pocscan_poc_file').multiselect({
        nonSelectedText: '选择POC文件',
        enableFiltering: true,
        enableCaseInsensitiveFiltering: true,
        includeFilterClearBtn: true,
        buttonWidth: '100%',
        maxHeight: 600,
        enableClickableOptGroups: true,
        enableCollapsibleOptGroups: true,
        filterPlaceholder: '搜索...',
        onChange: function (option, checked) {
            // 当选项状态改变时，重新排序选项
            sortSelectedOptionsToTop();
        }
    });
}

function process_form_data() {
    // 获取基础部分的字段值
    const id = $('#hidden_id').val();
    const name = $('#name').val();
    const description = $('#description').val();
    const sort_number = $('#sort_number').val();
    const status = $('#status').is(':checked');


    // 获取端口扫描部分的字段值
    const llmapi_enabled = $('#enable_llmapi').is(':checked');
    const qwen = $('#qwen').is(':checked');
    const kimi = $('#kimi').is(':checked');
    const deepseek = $('#deepseek').is(':checked');
    const icpPlus = $('#icpPlus').is(':checked');
    if (llmapi_enabled) {
        if (!qwen && !kimi && !deepseek && !icpPlus) {
            alert('请选择至少一种LLMAPI工具');
            return false;
        }
    }
    // 获取端口扫描部分的字段值
    const portscan_enabled = $('#enable_portscan').is(':checked');
    const nmap = $('#nmap').is(':checked');
    const masscan = $('#masscan').is(':checked');
    const gogo = $('#gogo').is(':checked');
    const portscanPort = $('#portscan_port').val();
    const portscanRate = $('#portscan_rate').val();
    const portscanTech = $('#portscan_tech').val();
    const portscanMaxOpenedPortPerIp = $('#portscan_PortscanConfig_maxOpenedPortPerIp').val();
    const portscanIsPing = $('#portscan_is_ping').is(':checked');
    if (portscan_enabled) {
        if (!nmap && !masscan && !gogo) {
            alert('请选择至少一种端口扫描工具');
            return false;
        }
        if (!isNotEmpty(portscanPort)) {
            alert('请输入端口扫描端口');
            return false;
        }
        if (!isNotEmpty(portscanRate)) {
            alert('请输入端口扫描速率');
            return false;
        }
        if (!isNotEmpty(portscanTech)) {
            alert('请输入端口扫描技术');
            return false;
        }
        if (!isNotEmpty(portscanMaxOpenedPortPerIp)) {
            alert('请输入最大打开端口数');
            return false;
        }
    }

    // 获取域名扫描部分的字段值
    const domainscan_enabled = $('#enable_domainscan').is(':checked');
    const massdns = $('#massdns').is(':checked');
    const subfinder = $('#subfinder').is(':checked');
    const domainscanIsIgnoreCdn = $('#domainscan_is_ignore_cdn').is(':checked');
    const domainscanIsIgnoreChinaOther = $('#domainscan_is_ignore_china_other').is(':checked');
    const domainscanIsIgnoreOutsideChina = $('#domainscan_is_ignore_outside_china').is(':checked');
    const domainscanIsIpPortScan = $('#domainscan_is_ip_port_scan').is(':checked');
    const domainscanIsIpSubnetPortScan = $('#domainscan_is_ip_subnet_port_scan').is(':checked');
    const domainscanMaxResolvedDomainPerIP = $('#domainscan_PortscanConfig_maxOpenedPortPerIp').val()
    const domainPortscanBin = $('#domain_portscan_bin').val();
    const domainPortscanPort = $('#domain_portscan_port').val()
    const domainPortscanRate = $('#domain_portscan_rate').val();
    const domainPortscanTech = $('#domain_portscan_tech').val();
    const domainPortscanMaxOpenedPortPerIp = $('#domain_portscan_PortscanConfig_maxOpenedPortPerIp').val();
    const domainscanWordlistFile = $('#domainscan_wordlist_file').val();
    if (domainscan_enabled) {
        if (!massdns && !subfinder) {
            alert('请选择至少一种域名扫描工具');
            return false;
        }
    }
    // 获取在线API部分的字段值
    const onlineapi_enabled = $('#enable_onlineapi').is(':checked');
    const fofa = $('#fofa').is(':checked');
    const hunter = $('#hunter').is(':checked');
    const quake = $('#quake').is(':checked');
    const whois = $('#whois').is(':checked');
    const icp = $('#icp').is(':checked');
    const onlineapiSearchByKeyword = $('#onlineapi_search_by_keyword').is(':checked');
    const onlineapiIsIgnoreCdn = $('#onlineapi_is_ignore_cdn').is(':checked');
    const onlineapiIsIgnoreChinaOther = $('#onlineapi_is_ignore_china_other').is(':checked');
    const onlineapiIsIgnoreOutsideChina = $('#onlineapi_is_ignore_outside_china').is(':checked');
    const onlineapiSearchStartTime = $('#onlineapi_search_start_time').val();
    const onlineapiSearchLimitCount = $('#onlineapi_search_limit_count').val();
    const onlineapiSearchPageSize = $('#onlineapi_search_page_size').val();
    if (onlineapi_enabled) {
        if (!fofa && !hunter && !quake && !whois && !icp) {
            alert('请选择至少一种在线API');
            return false;
        }
    }
    // 获取指纹识别部分的字段值
    const fingerprint_enabled = $('#enable_fingerprint').is(':checked');
    const fingerprintHttpx = $('#fingerprint_httpx').is(':checked');
    const fingerprintScreenshot = $('#fingerprint_screenshot').is(':checked');
    const fingerprintIconHash = $('#fingerprint_icon_hash').is(':checked');
    const fingerprintFingerprintx = $('#fingerprint_fingerprintx').is(':checked');
    if (fingerprint_enabled) {
        if (!fingerprintHttpx && !fingerprintFingerprintx) {
            alert('请选择至少一种指纹识别工具');
            return false;
        }
        if (fingerprintIconHash || fingerprintScreenshot) {
            if (!fingerprintHttpx) {
                alert('请选择Httpx工具，web指纹识别需要Httpx工具');
                return false;
            }
        }
    }
    // 获取POC扫描部分的字段值
    const pocscan_enabled = $('#enable_pocscan').is(':checked');
    const pocscan_bin = $('#pocscan_bin').val();
    const pocscan_poc_type = $('#pocscan_poc_type').val();
    const pocscan_base_web_status = $('#pocscan_base_web_status').is(':checked');
    const pocscanPocFile = $('#pocscan_poc_file').val().join(',');
    if (pocscan_enabled && pocscan_base_web_status) {
        if (!fingerprint_enabled || !fingerprintHttpx) {
            alert('请先开启指纹识别，并选择Httpx工具，web状态识别需要Httpx工具');
            return false;
        }
    }
    if (pocscan_enabled && pocscan_poc_type === "selectedPocFile") {
        if (pocscanPocFile.length === 0) {
            alert('请选择POC文件');
            return false;
        }
    }

    // 获取独立扫描部分的字段值
    const standalonePort = $('#standalone_port').val();
    const standaloneRate = $('#standalone_rate').val();
    const standaloneMaxOpenedPortPerIp = $('#standalone_PortscanConfig_maxOpenedPortPerIp').val();
    const standaloneIsPing = $('#standalone_is_ping').is(':checked');
    const standaloneIsVerbose = $('#standalone_is_verbose').is(':checked');
    const standaloneIsPocScan = $('#standalone_is_poc_scan').is(':checked');

    const configType = $('input[name="config_type"]:checked').val();
    // 组装成JavaScript对象
    return {
        id: id,
        name: name,
        description: description,
        status: status ? 'enable' : 'disable',
        sort_number: parseInt(sort_number),
        config_type: configType,
        llmapi: {
            enabled: llmapi_enabled,
            qwen: qwen,
            kimi: kimi,
            deepseek: deepseek,
            icpPlus: icpPlus,
        },
        portscan: {
            enabled: portscan_enabled,
            nmap: nmap,
            masscan: masscan,
            gogo: gogo,
            config: {
                port: portscanPort,
                rate: parseInt(portscanRate),
                ping: portscanIsPing,
                tech: portscanTech,
                maxOpenedPortPerIp: parseInt(portscanMaxOpenedPortPerIp)
            }
        },
        domainscan: {
            enabled: domainscan_enabled,
            massdns: massdns,
            subfinder: subfinder,
            config: {
                portscan: domainscanIsIpPortScan,
                subnetPortscan: domainscanIsIpSubnetPortScan,
                ignorecdn: domainscanIsIgnoreCdn,
                ignorechinaother: domainscanIsIgnoreChinaOther,
                ignoreoutsidechina: domainscanIsIgnoreOutsideChina,
                maxResolvedDomainPerIP: parseInt(domainscanMaxResolvedDomainPerIP),
                wordlistFile: domainscanWordlistFile,
                resultPortscanBin: domainPortscanBin,
                resultPortscanConfig: {
                    port: domainPortscanPort,
                    rate: parseInt(domainPortscanRate),
                    tech: domainPortscanTech,
                    maxOpenedPortPerIp: parseInt(domainPortscanMaxOpenedPortPerIp)
                }
            }
        },
        onlineapi: {
            enabled: onlineapi_enabled,
            fofa: fofa,
            hunter: hunter,
            quake: quake,
            whois: whois,
            icp: icp,
            config: {
                keywordsearch: onlineapiSearchByKeyword,
                ignorecdn: onlineapiIsIgnoreCdn,
                ignorechinaother: onlineapiIsIgnoreChinaOther,
                ignoreoutsidechina: onlineapiIsIgnoreOutsideChina,
                searchstarttime: onlineapiSearchStartTime,
                searchlimitcount: parseInt(onlineapiSearchLimitCount),
                searchpagesize: parseInt(onlineapiSearchPageSize)
            }
        },
        fingerprint: {
            enabled: fingerprint_enabled,
            config: {
                httpx: fingerprintHttpx,
                screenshot: fingerprintScreenshot,
                iconhash: fingerprintIconHash,
                fingerprintx: fingerprintFingerprintx
            }
        },
        pocscan: {
            enabled: pocscan_enabled,
            pocbin: pocscan_bin,
            config: {
                pocType: pocscan_poc_type,
                baseWebStatus: pocscan_base_web_status,
                pocFile: pocscanPocFile
            }
        },
        standalone: {
            config: {
                port: standalonePort,
                rate: parseInt(standaloneRate),
                ping: standaloneIsPing,
                verbose: standaloneIsVerbose,
                pocscan: standaloneIsPocScan,
                maxOpenedPortPerIp: parseInt(standaloneMaxOpenedPortPerIp)
            }
        }
    };
}

function fill_form_with_data(data) {
    // 基础部分
    $('#name').val(data.name);
    $('#description').val(data.description);
    $('#sort_number').val(data.sort_number);
    $('#status').prop('checked', data.status === 'enable');
    if (data.config_type === "staged") {
        // llmapi部分
        if (data.llmapi.enabled) {
            $('#qwen').prop('checked', data.llmapi.qwen);
            $('#kimi').prop('checked', data.llmapi.kimi);
            $('#deepseek').prop('checked', data.llmapi.deepseek);
            $('#icpPlus').prop('checked', data.llmapi.icpPlus);
        }
        // 端口扫描部分
        if (data.portscan.enabled) {
            $('#nmap').prop('checked', data.portscan.nmap);
            $('#masscan').prop('checked', data.portscan.masscan);
            $('#gogo').prop('checked', data.portscan.gogo);
            $('#portscan_port').val(data.portscan.config.port);
            $('#portscan_rate').val(data.portscan.config.rate);
            $('#portscan_tech').val(data.portscan.config.tech);
            $('#portscan_PortscanConfig_maxOpenedPortPerIp').val(data.portscan.config.maxOpenedPortPerIp);
            $('#portscan_is_ping').prop('checked', data.portscan.config.ping);
        }

        // 域名扫描部分
        if (data.domainscan.enabled) {
            $('#massdns').prop('checked', data.domainscan.massdns);
            $('#subfinder').prop('checked', data.domainscan.subfinder);
            $('#domainscan_is_ignore_cdn').prop('checked', data.domainscan.config.ignorecdn);
            $('#domainscan_is_ignore_china_other').prop('checked', data.domainscan.config.ignorechinaother);
            $('#domainscan_is_ignore_outside_china').prop('checked', data.domainscan.config.ignoreoutsidechina);
            $('#domainscan_is_ip_port_scan').prop('checked', data.domainscan.config.portscan);
            $('#domainscan_is_ip_subnet_port_scan').prop('checked', data.domainscan.config.subnetPortscan);
            $('#domainscan_PortscanConfig_maxResolvedDomainPerIP').val(data.domainscan.config.maxResolvedDomainPerIP);
            $('#domainscan_wordlist_file').val(data.domainscan.config.wordlistFile);
            if (data.domainscan.config.portscan || data.domainscan.config.subnetPortscan) {
                $('#domain_portscan_bin').val(data.domainscan.config.resultPortscanBin);
                $('#domain_portscan_port').val(data.domainscan.config.resultPortscanConfig.port);
                $('#domain_portscan_rate').val(data.domainscan.config.resultPortscanConfig.rate);
                $('#domain_portscan_tech').val(data.domainscan.config.resultPortscanConfig.tech);
                $('#domain_portscan_PortscanConfig_maxOpenedPortPerIp').val(data.domainscan.config.resultPortscanConfig.maxOpenedPortPerIp);
            }
        }
        // 在线API部分
        if (data.onlineapi.enabled) {
            $('#fofa').prop('checked', data.onlineapi.fofa);
            $('#hunter').prop('checked', data.onlineapi.hunter);
            $('#quake').prop('checked', data.onlineapi.quake);
            $('#whois').prop('checked', data.onlineapi.whois);
            $('#icp').prop('checked', data.onlineapi.icp);
            $('#onlineapi_search_by_keyword').prop('checked', data.onlineapi.config.searchbykeyword);
            $('#onlineapi_is_ignore_cdn').prop('checked', data.onlineapi.config.ignorecdn);
            $('#onlineapi_is_ignore_china_other').prop('checked', data.onlineapi.config.ignorechinaother);
            $('#onlineapi_is_ignore_outside_china').prop('checked', data.onlineapi.config.ignoreoutsidechina);
            $('#onlineapi_search_start_time').val(data.onlineapi.config.searchstarttime);
            $('#onlineapi_search_limit_count').val(data.onlineapi.config.searchlimitcount);
            $('#onlineapi_search_page_size').val(data.onlineapi.config.searchpagesize);
        }
        // 指纹识别部分
        if (data.fingerprint.enabled) {
            $('#fingerprint_httpx').prop('checked', data.fingerprint.config.httpx);
            $('#fingerprint_screenshot').prop('checked', data.fingerprint.config.screenshot);
            $('#fingerprint_icon_hash').prop('checked', data.fingerprint.config.iconhash);
            $('#fingerprint_fingerprintx').prop('checked', data.fingerprint.config.fingerprintx);
        }
        // POC扫描部分
        if (data.pocscan.enabled) {
            $('#pocscan_bin').val(data.pocscan.pocbin);
            $('#pocscan_base_web_status').prop('checked', data.pocscan.config.baseWebStatus);
            $('#pocscan_poc_type').val(data.pocscan.config.pocType);
            if (data.pocscan.config.pocType === "selectedPocFile") {
                $('#pocscan_poc_file').multiselect('enable');
                if (isNotEmpty(data.pocscan.config.pocFile)) {
                    let arr = data.pocscan.config.pocFile.split(',');
                    load_pocfile_data(arr);
                }
            } else {
                $('#pocscan_poc_file').multiselect('disable');
                load_pocfile_data(null);
            }
        }
    } else if (data.config_type === "standalone") {
        // 独立扫描部分
        $('#standalone_port').val(data.standalone.config.port);
        $('#standalone_rate').val(data.standalone.config.rate);
        $('#standalone_PortscanConfig_maxOpenedPortPerIp').val(data.standalone.config.maxOpenedPortPerIp);
        $('#standalone_is_ping').prop('checked', data.standalone.config.ping);
        $('#standalone_is_verbose').prop('checked', data.standalone.config.verbose);
        $('#standalone_is_poc_scan').prop('checked', data.standalone.config.pocscan);
    }
    $('#enable_llmapi').prop('checked', data.llmapi.enabled);
    $('#enable_portscan').prop('checked', data.portscan.enabled);
    $('#enable_domainscan').prop('checked', data.domainscan.enabled);
    $('#enable_onlineapi').prop('checked', data.onlineapi.enabled);
    $('#enable_fingerprint').prop('checked', data.fingerprint.enabled);
    $('#enable_pocscan').prop('checked', data.pocscan.enabled);
    $('#llmapi_config input, #llmapi_config select').prop('disabled', !data.llmapi.enabled);
    $('#portscan_config input, #portscan_config select').prop('disabled', !data.portscan.enabled);
    $('#domainscan_config input, #domainscan_config select').prop('disabled', !data.domainscan.enabled);
    $('#onlineapi_config input, #onlineapi_config select').prop('disabled', !data.onlineapi.enabled);
    $('#fingerprint_config input, #fingerprint_config select').prop('disabled', !data.fingerprint.enabled);
    $('#pocscan_config input, #pocscan_config select').prop('disabled', !data.pocscan.enabled);

    if (data.config_type === "staged") {
        $('#staged-tab').tab('show');
        $('input[name="config_type"][value="staged"]').prop('checked', true);
    } else if (data.config_type === "standalone") {
        $('#standalone-tab').tab('show');
        $('input[name="config_type"][value="standalone"]').prop('checked', true);
    }
}

function generate_profile_name() {
    const stringBuilder = [];
    const configType = $('input[name="config_type"]:checked').val();
    if (configType === "staged") {
        if ($('#enable_llmapi').is(':checked')) {
            if ($('#qwen').is(':checked') || $('#kimi').is(':checked') || $('#deepseek').is(':checked')) {
                stringBuilder.push("LLMAPI");
            }
            if ($('#icpPlus').is(':checked')) {
                stringBuilder.push("ICPPlus");
            }
        }
        if ($('#enable_portscan').is(':checked')) {
            const port = $('#portscan_port').val();
            if (port === "--top-ports 1000") stringBuilder.push("端口扫描(Top1000)");
            else if (port === "--top-ports 100") stringBuilder.push("端口扫描(Top100)");
            else if (port === "--top-ports 10") stringBuilder.push("端口扫描(Top10)");
            else if (port === "1-65535") stringBuilder.push("端口扫描(全端口)");
            else stringBuilder.push("端口扫描(自定义)");
        }
        if ($('#enable_domainscan').is(':checked')) {
            stringBuilder.push("域名扫描");
            if ($('#domainscan_wordlist_file').val() === "medium") {
                stringBuilder.push("中型字典");
            }
            if ($('#domainscan_is_ip_port_scan').is(':checked') || $('#domainscan_is_ip_subnet_port_scan').is(':checked')) {
                stringBuilder.push("端口扫描");
            }
        }
        if ($('#enable_onlineapi').is(':checked')) {
            stringBuilder.push("在线API");
            if ($('#onlineapi_search_by_keyword').is(':checked')) {
                stringBuilder.push("关键字搜索");
            }
        }
        if ($('#enable_fingerprint').is(':checked')) {
            stringBuilder.push("指纹识别");
        }
        if ($('#enable_pocscan').is(':checked')) {
            stringBuilder.push("Poc扫描");
            //stringBuilder.push($('#pocscan_bin').val());
            if ($('#pocscan_poc_type').val() === "matchFinger") {
                stringBuilder.push("指纹匹配");
            } else {
                if ($('#pocscan_base_web_status').is(':checked')) {
                    stringBuilder.push("HTTP状态码");
                }
            }
        }
    } else {
        stringBuilder.push("(SA)");
        const port = $('#standalone_port').val();
        if (port === "--top-ports 1000") stringBuilder.push("独立任务(Top1000)");
        else if (port === "--top-ports 100") stringBuilder.push("独立任务(Top100)");
        else if (port === "--top-ports 10") stringBuilder.push("独立任务(Top10)");
        else if (port === "1-65535") stringBuilder.push("独立任务(全端口)");
        else if (port === "top1") stringBuilder.push("独立任务(top1)");
        else stringBuilder.push("独立任务(自定义)");
    }
    return stringBuilder.join("-");
}

function load_profile_data(id) {
    $.ajax({
        type: 'POST',
        url: '/profile-get',
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


function save_profile_data() {
    const profileInfoData = process_form_data();
    if (profileInfoData === false) return;
    $.ajax({
        type: 'POST',
        url: '/profile-save',
        contentType: 'application/json',
        data: JSON.stringify(profileInfoData),
        success: function (data) {
            if (data.status === 'success') {
                // 成功，更新隐藏的ID字段并弹出成功提示
                if (profileInfoData.id === "") {
                    $('#hidden_id').val(data['msg']);
                }
                show_response_message("保存成功", data, function () {
                })
            } else {
                swal('失败', data.message, 'error');
            }
        },
        error: function (xhr, status, error) {
            swal('失败', '请求失败: ' + error, 'error');
        }
    });
}
