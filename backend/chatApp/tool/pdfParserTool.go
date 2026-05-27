package tool

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	pdfParser "github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// PDFToTextRequest 大模型调用工具的入参结构体（明确参数要求）
type PDFToTextRequest struct {
	FilePath string `json:"file_path" jsonschema:"required,description=本地PDF文件的绝对路径（例如：D:\\test\\document.pdf 或 /home/user/document.pdf）"`
	ToPages  bool   `json:"to_pages" jsonschema:"default=false,description=是否按页面分割文本（true=分页输出，false=合并所有页为一个文本，默认false）"`
}

// PDFToTextResult 工具返回的结构化结果（大模型可直接解析）
type PDFToTextResult struct {
	Success    bool                   `json:"success" jsonschema:"description=解析是否成功"`
	Content    string                 `json:"content,omitempty" jsonschema:"description=合并后的纯文本（ToPages=false时返回）"`
	Pages      []PDFPageText          `json:"pages,omitempty" jsonschema:"description=分页文本（ToPages=true时返回）"`
	TotalPages int                    `json:"total_pages" jsonschema:"description=PDF总页数"`
	ErrorMsg   string                 `json:"error_msg,omitempty" jsonschema:"description=错误信息（失败时返回）"`
	Meta       map[string]interface{} `json:"meta,omitempty" jsonschema:"description=元数据（方便追溯）"`
}

// PDFPageText 单页文本结构（分页模式下使用）
type PDFPageText struct {
	PageNum int    `json:"page_num" jsonschema:"description=页码（从1开始）"`
	Content string `json:"content" jsonschema:"description=单页纯文本"`
}

// ConvertPDFToText 核心逻辑：PDF转纯文本（工具执行入口）
func ConvertPDFToText(ctx context.Context, req *PDFToTextRequest) (*PDFToTextResult, error) {
	result := PDFToTextResult{
		Meta: map[string]interface{}{
			"file_path":  req.FilePath,
			"to_pages":   req.ToPages,
			"parse_time": time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	// 1. 参数校验（必填参数检查）
	if req.FilePath == "" {
		result.Success = false
		result.ErrorMsg = "参数错误：必须传入 file_path（本地PDF文件的绝对路径）"
		return &result, errors.New(result.ErrorMsg)
	}

	// 2. 打开本地PDF文件
	file, err := os.Open(req.FilePath)
	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("打开PDF文件失败：%v（请检查路径是否正确、文件是否存在）", err)
		return &result, errors.New(result.ErrorMsg)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("close file failed: %v", err)
		}
	}(file)

	// 3. 初始化Eino PDF解析器（无超时配置，极简核心）
	pdfIns, err := pdfParser.NewPDFParser(ctx, &pdfParser.Config{
		ToPages: req.ToPages, // 按大模型传入的参数决定是否分页
	})
	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("初始化PDF解析器失败：%v", err)
		return &result, errors.New(result.ErrorMsg)
	}

	// 4. 核心：解析PDF为纯文本
	docs, err := pdfIns.Parse(ctx, file,
		parser.WithURI(req.FilePath),
		parser.WithExtraMeta(result.Meta),
	)
	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("PDF解析失败：%v（仅支持文本型PDF，不支持扫描件/加密PDF）", err)
		return &result, errors.New(result.ErrorMsg)
	}

	// 5. 构造成功结果
	result.Success = true
	result.TotalPages = len(docs)

	if req.ToPages {
		// 分页模式：按页码整理文本（用索引+1作为页码，可靠无依赖）
		pages := make([]PDFPageText, 0, len(docs))
		for idx, doc := range docs {
			pages = append(pages, PDFPageText{
				PageNum: idx + 1,
				Content: doc.Content,
			})
		}
		result.Pages = pages
	} else {
		// 合并模式：拼接所有页文本
		var contentBuilder string
		for _, doc := range docs {
			contentBuilder += doc.Content + "\n"
		}
		result.Content = contentBuilder
	}

	// 6. 结果序列化为JSON（大模型可直接解析）
	return &result, nil
}

func CreatePDFToTextTool() tool.InvokableTool {

	pdfTool, err := utils.InferTool(
		"pdf_to_text",
		"将本地PDF文件转换为纯文本，仅支持文本型PDF（可复制文字），不支持扫描件、加密PDF。需传入本地PDF的绝对路径，可选择按页面分割或合并所有页。",
		ConvertPDFToText,
	)
	if err != nil {
		log.Fatalf("invoke pdf_to_text failed: %v", err)
	}

	fmt.Println("PDF转纯文本工具初始化完成（大模型可调用）")

	return pdfTool

}