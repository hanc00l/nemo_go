$(function () {
    load_worker_config();
    $('#buttonSaveAPI').click(function () {
        var data = {
            api: {
                searchPageSize: parseInt($('#input_pagesize').val()),
                searchLimitCount: parseInt($('#input_limitcount').val()),
                fofa: {
                    key: $('#input_fofa_key').val()
                },
                hunter: {
                    key: $('#input_hunter_key').val()
                },
                quake: {
                    key: $('#input_quake_key').val()
                },
                icp: {
                    key: $('#input_icp_key').val()
                },
                icpPlus: {
                    key: $('#input_icpPlus_key').val()
                }
            },
        };
        save_worker_config(data);
    });
    $("#buttonTestAPI").click(function () {
        test_online_api();
    });
    $('#buttonTestLLM').click(function () {
        test_llm_api();
    })
    $('#buttonSaveLLM').click(function () {
        var data = {
            llmapi: {
                kimi: {
                    api: $('#input_kimi_api').val(),
                    model: $('#input_kimi_model').val(),
                    token: $('#input_kimi_token').val()
                },
                deepseek: {
                    api: $('#input_deepseek_api').val(),
                    model: $('#input_deepseek_model').val(),
                    token: $('#input_deepseek_token').val()
                },
                qwen: {
                    api: $('#input_qwen_api').val(),
                    model: $('#input_qwen_model').val(),
                    token: $('#input_qwen_token').val()
                }
            }
        };
        save_worker_config(data);
    });
    $('#buttonSaveProxy').click(function () {
        var data = {
            proxy: {
                host: $('#text_proxy_list').val().split('\n')
            }
        };
        save_worker_config(data);
    });
});

function save_worker_config(data) {
    $.ajax({
        type: 'POST',
        url: '/config-worker-save-config',
        data: JSON.stringify(data),
        dataType: 'json',
        success: function (response) {
            show_response_message("保存成功", response, function () {
            });
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function load_worker_config() {
    $.ajax({
        type: 'POST',
        url: '/config-worker-load-config',
        dataType: 'json',
        success: function (response) {
            $('#input_pagesize').val(response.api.searchPageSize);
            $('#input_limitcount').val(response.api.searchLimitCount);
            $('#input_fofa_key').val(response.api.fofa.key);
            $('#input_hunter_key').val(response.api.hunter.key);
            $('#input_quake_key').val(response.api.quake.key);
            $('#input_icp_key').val(response.api.icp.key);
            $('#input_icpPlus_key').val(response.api.icpPlus.key);
            $('#input_kimi_api').val(response.llmapi.kimi.api);
            $('#input_kimi_model').val(response.llmapi.kimi.model);
            $('#input_kimi_token').val(response.llmapi.kimi.token);
            $('#input_deepseek_api').val(response.llmapi.deepseek.api);
            $('#input_deepseek_model').val(response.llmapi.deepseek.model);
            $('#input_deepseek_token').val(response.llmapi.deepseek.token);
            $('#input_qwen_api').val(response.llmapi.qwen.api);
            $('#input_qwen_model').val(response.llmapi.qwen.model);
            $('#input_qwen_token').val(response.llmapi.qwen.token);
            $('#text_proxy_list').val(response.proxy.host.join('\n'));
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function test_online_api() {
    $.ajax({
        type: 'POST',
        url: '/config-worker-test-onlineapi',
        dataType: 'json',
        success: function (data) {
            if (data['status'] === "success") {
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
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function test_llm_api() {
    $.ajax({
        type: 'POST',
        url: '/config-worker-test-llmapi',
        dataType: 'json',
        success: function (data) {
            if (data['status'] === "success") {
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
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}