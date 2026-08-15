# Argus 观枢

Argus（观枢）是一个 Kubernetes 原生 AI 大模型出口安全网关。在业务 Pod 零改造的前提下，通过集群全局透明出口引流，对 LLM 出站流量做协议识别、Prompt 提取、多检测器流水线、风险打分与放行/阻断决策，并产出标准化安全事件 AIEvent。

## 构建命令

    make build

## 生成 protobuf 桩代码

    make proto

## 运行测试

    make test

## 整理依赖

    make tidy
