// Package semantic 实现基于语义向量的检测能力，通过 Embedding 模型将 Prompt/Response
// 向量化后与恶意样本库做余弦相似度比对，覆盖零日 Jailbreak、隐晦 Prompt Injection
// 等规则难以覆盖的场景。依赖可选的向量数据库或本地 Faiss 索引。
package semantic
