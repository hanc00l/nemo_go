$(function () {
    const workspace_select = $('#switch_workspace');
    load_select_list("/dashboard-workspace-list", workspace_select, $('#current_workspace').val());
    load_asset_count();
    load_asset_charts();
    load_vul_charts();
    load_task_charts();
    load_user_charts();
    workspace_select.on('change', function () {
        change_workspace();
    });
});

function change_workspace() {
    let workspace_id = $('#switch_workspace').val();
    if (workspace_id === "") {
        alert("请选择一个工作空间");
        return;
    }
    $.ajax({
        type: 'POST',
        url: "/dashboard-workspace-change",
        data: {workspaceId: workspace_id},
        dataType: 'json',
        success: function (response) {
            show_response_message("切换工作空间成功", response, function () {
                window.location.reload();
            })
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function load_asset_count() {
    const workspaceId = $('#current_workspace').val();
    if (workspaceId === "") {
        return;
    }
    $.ajax({
        type: 'POST',
        url: '/dashboard-asset-count',
        data: {workspaceId: workspaceId},
        dataType: 'json',
        success: function (data) {
            $('#asset-count').text(data.total);
            $('#ip-count').text(data.ip);
            $('#domain-count').text(data.domain);
            $('#vul-count').text(data.vul);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function load_asset_charts() {
    const workspaceId = $('#current_workspace').val();
    if (workspaceId === "") {
        return;
    }
    $.ajax({
        type: 'POST',
        url: '/dashboard-asset-chart',
        data: {workspaceId: workspaceId},
        dataType: 'json',
        success: function (data) {
            init_asset_vul_chart_data(document.querySelector('#asset_chart'), data, '近期资产更新数量');
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}


function load_vul_charts() {
    const workspaceId = $('#current_workspace').val();
    if (workspaceId === "") {
        return;
    }
    $.ajax({
        type: 'POST',
        url: '/dashboard-vul-chart',
        data: {workspaceId: workspaceId},
        dataType: 'json',
        success: function (data) {
            init_asset_vul_chart_data(document.querySelector('#vul_chart'), data, '近期漏洞更新数量');
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}


function load_task_charts() {
    const workspaceId = $('#current_workspace').val();
    if (workspaceId === "") {
        return;
    }
    $.ajax({
        type: 'POST',
        url: '/dashboard-task-chart',
        data: {workspaceId: workspaceId},
        dataType: 'json',
        success: function (data) {
            init_task_user_chart_data(document.querySelector('#task_chart'), data,"近期任务数量");
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function load_user_charts() {
    $.ajax({
        type: 'POST',
        url: '/dashboard-user-chart',
        dataType: 'json',
        success: function (data) {
            init_task_user_chart_data(document.querySelector('#user_chart'), data,"近期用户登录数量");
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function init_asset_vul_chart_data(chartObj, data, title) {
    const options = {
        series: data.series,
        chart: {
            height: 350,
            type: 'line',
            dropShadow: {
                enabled: true,
                color: '#000',
                top: 18,
                left: 7,
                blur: 10,
                opacity: 0.5
            },
            zoom: {
                enabled: false
            },
            toolbar: {
                show: false
            }
        },
        colors: ['#F64E60',  '#FFC107', '#00E396', '#0077FF'],
        dataLabels: {
            enabled: true,
        },
        stroke: {
            curve: 'smooth'
        },
        title: {
            text: title,
            align: 'left'
        },
        grid: {
            borderColor: '#e7e7e7',
            row: {
                colors: ['#f3f3f3', 'transparent'], // takes an array which will be repeated on columns
                opacity: 0.5
            },
        },
        markers: {
            size: 1
        },
        xaxis: {
            categories: data.categories,
            title: {
                text: '日期'
            }
        },
        yaxis: {
            title: {
                text: '数量'
            },
        },
        legend: {
            position: 'top',
            horizontalAlign: 'right',
            floating: true,
            offsetY: -25,
            offsetX: -5
        }
    };

    const chart = new ApexCharts(chartObj, options);
    chart.render();
}

function init_task_user_chart_data(chartObj, data,title) {
    const options = {
        series: data.series,
        chart: {
            type: 'bar',
            height: 350
        },
        plotOptions: {
            bar: {
                horizontal: false,
                columnWidth: '55%',
                borderRadius: 5,
                borderRadiusApplication: 'end',
                dataLabels: {
                    position: 'top'
                }
            },
        },
        dataLabels: {
            enabled: true,
        },
        stroke: {
            show: true,
            width: 2,
            colors: ['transparent']
        },
        title: {
            text: title,
            align: 'left'
        },
        xaxis: {
            categories: data.categories,
        },
        yaxis: {
            title: {
                text: '数量'
            }
        },
        fill: {
            opacity: 1
        },
        tooltip: {
            y: {
                formatter: function (val) {
                    return "总计" + val + "个任务"
                }
            }
        }
    };

    const chart = new ApexCharts(chartObj, options);
    chart.render();

}