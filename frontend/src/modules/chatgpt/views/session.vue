<template>
	<cl-crud ref="Crud">
		<cl-row>
			<!-- 刷新按钮 -->
			<cl-refresh-btn />
			<!-- 新增按钮 -->
			<cl-add-btn />
			<!-- 删除按钮 -->
			<cl-multi-delete-btn />
			<el-button type="primary" @click="DialogData['addAccountsVisible'] = true"
				>批量添加</el-button
			>
			<el-button type="primary" @click="DialogData['showLogVisible'] = true"
				>查看日志</el-button
			>
			<cl-flex1 />
			<!-- 关键字搜索 -->
			<cl-search-key />
		</cl-row>

		<cl-row>
			<!-- 数据表格 -->
			<cl-table ref="Table" />
		</cl-row>

		<cl-row>
			<cl-flex1 />
			<!-- 分页控件 -->
			<cl-pagination />
		</cl-row>

		<!-- 新增、编辑 -->
		<cl-upsert ref="Upsert" />
	</cl-crud>
	<f-k-arkos
		:public-key="publicKey"
		mode="lightbox"
		arkosUrl=""
		@onCompleted="onCompleted($event)"
		@onError="onError($event)"
	/>
	<cl-dialog title="添加账号" v-model="DialogData['addAccountsVisible']">
		<div>
			<el-input
				v-model="DialogData['accounts']"
				:rows="15"
				type="textarea"
				placeholder="复制账号密码到这里,每行一个,格式: 邮箱,密码"
			/>
			<el-button type="primary" @click="submitAccounts">提交</el-button>
		</div>
	</cl-dialog>
	<cl-dialog
		title="添加账号日志"
		v-model="DialogData['showLogVisible']"
		@open="init"
		@close="closeSocket"
		:destroy-on-close="true"
		:close-on-click-modal="false"
	>
		<div id="terminal"></div>
	</cl-dialog>
</template>

<script lang="ts" name="chatgpt-session" setup>
import { useCrud, useTable, useUpsert } from "@cool-vue/crud";
import { useCool } from "/@/cool";

const { service } = useCool();

// cl-upsert 配置
const Upsert = useUpsert({
	items: [
		{ label: "邮箱", prop: "email", required: true, component: { name: "el-input" } },
		{ label: "密码", prop: "password", required: true, component: { name: "el-input" } },
		{
			label: "状态",
			prop: "status",
			component: {
				name: "el-switch",
				props: {
					activeValue: 1,
					inactiveValue: 0
				}
			}
		},
		{
			label: "PLUS",
			prop: "isPlus",
			component: {
				name: "el-switch",
				props: {
					activeValue: 1,
					inactiveValue: 0
				}
			}
		},
		{
			label: "session",
			prop: "officialSession",
			component: { name: "el-input", props: { type: "textarea", rows: 4 } }
		},
		{
			label: "备注",
			prop: "remark",
			component: { name: "el-input", props: { type: "textarea", rows: 4 } }
		}
	],
	onOpened(data) {
		// // 自动生成uuid 作为userToken
		// if (!data.userID) {
		// 	data.userID = 0;
		// }
		localStorage.removeItem("arkoseToken");

		if (!data.officialSession) {
			ElMessage({
				message: "请稍等,人机验证进行中.",
				type: "warning"
			});
			window.myEnforcement.run();
		}
	},
	onSubmit(data, { done, close, next }) {
		let arkoseToken = localStorage.getItem("arkoseToken");
		let w = window;
		if (arkoseToken) {
			localStorage.removeItem("arkoseToken");

			next({ ...data, arkoseToken });
			done();
			close();
		} else {
			if (!data.officialSession) {
				w.myEnforcement.run();
				ElMessage({
					message: "请稍等,人机验证进行中,验证完成后请重新点击确定保存.",
					type: "warning"
				});
				// alert("请先完成人机验证");

				done();
			} else {
				next(data);
				done();
				close();
			}
		}
	}
});

// cl-table 配置
const Table = useTable({
	columns: [
		{ type: "selection" },
		{ label: "id", prop: "id" },
		{ label: "创建时间", prop: "createTime", sortable: true },
		{ label: "更新时间", prop: "updateTime", sortable: true },
		{ label: "邮箱", prop: "email", sortable: true },
		{ label: "密码", prop: "password", sortable: true },
		{ label: "状态", prop: "status", component: { name: "cl-switch" }, sortable: true },
		{ label: "PLUS", prop: "isPlus", component: { name: "cl-switch" }, sortable: true },
		{
			label: "session",
			prop: "officialSession",
			showOverflowTooltip: true,
			sortable: true
		},
		{ label: "备注", prop: "remark", showOverflowTooltip: true, sortable: true },
		{ type: "op", buttons: ["edit", "delete"] }
	]
});

// cl-crud 配置
const Crud = useCrud(
	{
		service: service.chatgpt.session
	},
	(app) => {
		app.refresh();
	}
);
</script>
<script lang="ts">
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";

const { service } = useCool();

import FKArkos from "./FKArkos.vue";
import { ElMessage, ElMessageBox } from "element-plus";

import { defineComponent } from "vue";
export default defineComponent({
	components: {
		FKArkos
	},
	data() {
		return {
			// publicKey: process.env.VUE_APP_ARKOSE_PUBLIC_KEY,
			publicKey: "0A1D34FC-659D-4E23-B17B-694DCFCF6A6C",
			arkoseToken: "",
			DialogData: {
				accounts: "",
				addAccountsVisible: false,
				showLogVisible: false
			},
			socketData: {
				term: null,
				socket: null,
				lockReconnect: false, //是否真正建立连接
				timeout: 28 * 1000, //30秒一次心跳
				timeoutObj: null, //心跳心跳倒计时
				serverTimeoutObj: null, //心跳倒计时
				timeoutnum: null //断开 重连倒计时
			}
		};
	},
	beforeUnmount() {
		this.closeSocket();
	},
	methods: {
		onCompleted(token: string) {
			console.log("onCompleted---------->", token);
			ElMessage({
				message: "人机验证已完成.",
				type: "success"
			});
			localStorage.setItem("arkoseToken", token);
			// 设置过期时间 tokenExpire 为4分钟
			// let tokenExpire = now.getTime() + 4 * 60 * 1000;
			// localStorage.setItem("tokenExpire", tokenExpire);

			this.arkoseToken = token;
			// router.replace({ path: "/dashboard" });
		},
		onError(errorMessage: any) {
			// alert(errorMessage);
			ElMessageBox.alert("加载人机验证失败,请刷新页面重试!", errorMessage.error.error, {
				// if you want to disable its autofocus
				// autofocus: false,
				confirmButtonText: "OK"
				// callback: (action: Action) => {
				// 	ElMessage({
				// 		type: "info",
				// 		message: `action: ${action}`
				// 	});
				// }
			});
		},

		onSubmit() {
			if (!this.arkoseToken) {
				window.myEnforcement.run();
			}
		},
		submitAccounts() {
			let accounts = this.DialogData["accounts"];
			if (!accounts) {
				ElMessage({
					message: "请输入账号密码",
					type: "warning"
				});
				return;
			}
			service.chatgpt.session
				.addbulk({
					accounts
				})
				.then((res) => {
					ElMessage({
						message: "提交成功",
						type: "success"
					});
					console.log("res", res);
					this.DialogData["addAccountsVisible"] = false;
					this.DialogData["accounts"] = "";
					this.DialogData["showLogVisible"] = true;
				})
				.catch((err) => {
					console.log("err", err);
					ElMessage.error(err);
					// this.DialogData["showLogVisible"] = true;
				});
		},

		closeSocket() {
			if (this.socketData["socket"]) {
				this.socketData["socket"].close();
				this.socketData["socket"] = null;
			}
		},
		socketReconnect() {
			//重新连接
			if (this.socketData["lockReconnect"]) {
				return;
			}
			this.socketData["lockReconnect"] = true;
			//没连接上会一直重连，设置延迟避免请求过多
			this.socketData["timeoutnum"] && clearTimeout(this.socketData["timeoutnum"]);
			this.socketData["timeoutnum"] = setTimeout(function () {
				this.init();
				this.socketData["lockReconnect"] = false;
			}, 2000);
		},
		socketReset() {
			clearTimeout(this.socketData["timeoutObj"]);
			clearTimeout(this.socketData["serverTimeoutObj"]);
			this.socketStart();
		},
		socketStart() {
			this.socketData["timeoutObj"] && clearTimeout(this.socketData["timeoutObj"]);
			this.socketData["serverTimeoutObj"] &&
				clearTimeout(this.socketData["serverTimeoutObj"]);
			this.socketData["timeoutObj"] = setTimeout(function () {
				if (this.socket.readyState == 1) {
					this.socket.send("ping");
				} else {
					this.socketReconnect();
				}
				this.socketData["serverTimeoutObj"] = setTimeout(function () {
					this.close();
				}, this.socketData["timeout"]);
			}, this.socketData["timeout"]);
		},
		init() {
			// 实例化socket
			this.socketData["socket"] = new WebSocket(`/socket`);
			// 监听socket连接
			this.socketData["socket"].onopen = this.open;
			// 监听socket错误信息
			this.socketData["socket"].onerror = this.socketReconnect;
			// 监听socket消息
			this.socketData["socket"].onmessage = this.getMessage;
			// 发送socket消息
			this.socketData["socket"].onsend = this.send;
		},
		open() {
			this.initXterm();
			console.log("socket连接成功");
			this.socketStart();
		},
		getMessage(msg) {
			//msg是返回的数据
			let msgs = JSON.parse(msg.data);
			this.socketData["socket"].send(
				JSON.stringify({
					e: "ping"
				})
			); //有事没事ping一下，看看ws还活着没
			// console.log("getMessage", msgs, msgs["d"]);

			if (msgs["e"] !== "ping") {
				if (
					msgs["d"] &&
					typeof msgs["d"] === "object" &&
					msgs["d"].constructor === Object
				) {
					this.socketData["term"].write(`${JSON.stringify(msgs["d"])}\r \n`);
				} else {
					this.socketData["term"].write(`${msgs["d"]}\r \n`);
				}
			}
			//收到服务器信息，心跳重置
			this.socketReset();
		},
		send(order) {
			this.socketData["socket"].send(order);
		},
		initXterm() {
			if (this.socketData["term"]) {
				this.socketData["term"].dispose();
			}

			this.socketData["term"] = new Terminal({
				disableStdin: false
			});
			// 	{
			// 	rendererType: "canvas", //渲染类型
			// 	rows: 35, //行数
			// 	convertEol: true, //启用时，光标将设置为下一行的开头
			// 	scrollback: 10, //终端中的回滚量
			// 	disableStdin: false, //是否应禁用输入
			// 	cursorStyle: "underline", //光标样式
			// 	cursorBlink: true, //光标闪烁
			// 	theme: {
			// 		foreground: "yellow", //字体
			// 		background: "#060101", //背景色
			// 		cursor: "help" //设置光标
			// 	}
			// }
			this.socketData["term"].open(document.getElementById("terminal"));
			let fitAddon = new FitAddon();
			// this.socketData["term"].loadAddon(fitAddon);
			fitAddon.activate(this.socketData["term"]);
			this.socketData["term"].write(`\r \n`);
		}
	}
});
</script>
