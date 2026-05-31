package storypack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"namelesswatch/internal/roleplay"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

const (
	defaultStoryTitle   = "未命名规则怪谈"
	defaultInitialScene = "entrance"
)

var scaffoldDirs = []string{"photo", "bgm"}

var scaffoldFiles = []scaffoldFile{
	{Path: "metadata.json", Content: metadataJSON},
	{Path: "briefing.json", Content: briefingJSON},
	{Path: "achievements.json", Content: achievementsJSON},
	{Path: "scene.md", Content: sceneMarkdown},
	{Path: "rule.md", Content: ruleMarkdown},
	{Path: "true.md", Content: trueMarkdown},
	{Path: "memory.md", Content: memoryMarkdown},
	{Path: "endings.md", Content: endingsMarkdown},
	{Path: "photo/metadata.json", Content: photoMetadataJSON},
	{Path: "bgm/metadata.json", Content: bgmMetadataJSON},
}

var requiredPackFiles = []string{
	"metadata.json",
	"briefing.json",
	"achievements.json",
	"scene.md",
	"rule.md",
	"true.md",
	"memory.md",
	"endings.md",
	"photo/metadata.json",
	"bgm/metadata.json",
}

func ScaffoldFilePaths() []string {
	paths := make([]string, 0, len(scaffoldFiles))
	for _, file := range scaffoldFiles {
		paths = append(paths, file.Path)
	}
	return paths
}

func RequiredPackFiles() []string {
	return slices.Clone(requiredPackFiles)
}

func SafeFolderName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultStoryTitle
	}
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	value = strings.TrimSpace(replacer.Replace(value))
	if value == "" || value == "." || value == ".." {
		return defaultStoryTitle
	}
	return value
}

type ScaffoldOptions struct {
	Title        string
	InitialScene string
	Force        bool
}

type scaffoldFile struct {
	Path    string
	Content func(ScaffoldOptions) string
}

type ValidationReport struct {
	Root     string
	Title    string
	Problems []string
	Warnings []string
}

type metadataTemplate struct {
	Title          string               `json:"title"`
	InitialScene   string               `json:"initialScene"`
	ScenePositions map[string][]float64 `json:"scenePositions"`
}

type briefingTemplate struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Items       []briefingItem `json:"items"`
	ConfirmText string         `json:"confirmText"`
}

type briefingItem struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type bgmMetadataTemplate struct {
	Tracks        map[string]bgmTrackTemplate `json:"tracks"`
	SceneDefaults map[string]string           `json:"sceneDefaults"`
}

type bgmTrackTemplate struct {
	Name string `json:"name"`
	File string `json:"file"`
}

func ScaffoldPack(root string, opts ScaffoldOptions) ([]string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve target path: %w", err)
	}
	opts = normalizeScaffoldOptions(absRoot, opts)

	if err := ensureRootDir(absRoot); err != nil {
		return nil, err
	}
	if !opts.Force {
		conflicts, err := existingScaffoldFiles(absRoot)
		if err != nil {
			return nil, err
		}
		if len(conflicts) > 0 {
			return nil, fmt.Errorf("target already contains scaffold files: %s (enable overwrite to replace them)", strings.Join(conflicts, ", "))
		}
	}

	for _, dir := range scaffoldDirs {
		if err := os.MkdirAll(filepath.Join(absRoot, filepath.FromSlash(dir)), 0o755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	written := make([]string, 0, len(scaffoldFiles))
	for _, file := range scaffoldFiles {
		target := filepath.Join(absRoot, filepath.FromSlash(file.Path))
		if err := writeFileAtomic(target, []byte(file.Content(opts)), 0o600, opts.Force); err != nil {
			return written, err
		}
		written = append(written, file.Path)
	}
	return written, nil
}

func ValidatePack(root string) (ValidationReport, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ValidationReport{}, fmt.Errorf("resolve target path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return ValidationReport{}, fmt.Errorf("stat target path: %w", err)
	}
	if !info.IsDir() {
		return ValidationReport{}, fmt.Errorf("target path is not a directory: %s", absRoot)
	}

	files, err := collectPackFiles(absRoot)
	if err != nil {
		return ValidationReport{}, err
	}

	report := ValidationReport{Root: absRoot}
	for _, dir := range scaffoldDirs {
		dirPath := filepath.Join(absRoot, filepath.FromSlash(dir))
		dirInfo, err := os.Stat(dirPath)
		if errors.Is(err, os.ErrNotExist) {
			report.Problems = append(report.Problems, fmt.Sprintf("missing directory: %s", dir))
			continue
		}
		if err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("cannot read directory %s: %v", dir, err))
			continue
		}
		if !dirInfo.IsDir() {
			report.Problems = append(report.Problems, fmt.Sprintf("%s is not a directory", dir))
		}
	}

	for _, fileName := range requiredPackFiles {
		if _, ok := files[normalizePackPath(fileName)]; !ok {
			report.Problems = append(report.Problems, fmt.Sprintf("missing file: %s", fileName))
		}
	}

	if raw, ok := files[normalizePackPath("metadata.json")]; ok {
		validateMetadata(raw, &report)
	}
	if raw, ok := files[normalizePackPath("briefing.json")]; ok {
		validateBriefing(raw, &report)
	}
	if raw, ok := files[normalizePackPath("photo/metadata.json")]; ok {
		validatePhotoMetadata(raw, &report)
	}
	if raw, ok := files[normalizePackPath("bgm/metadata.json")]; ok {
		validateBGMMetadata(raw, &report)
	}
	for _, fileName := range roleplay.RequiredStoryFiles {
		if raw, ok := files[normalizePackPath(fileName)]; ok && strings.TrimSpace(raw) == "" {
			report.Problems = append(report.Problems, fmt.Sprintf("%s must not be empty", fileName))
		}
	}

	if len(report.Problems) == 0 {
		game, importReport, err := roleplay.NewLibraryGame(files)
		if err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("core pack parse failed: %v", err))
		} else if len(importReport.Missing) > 0 {
			report.Problems = append(report.Problems, fmt.Sprintf("core pack missing files: %s", strings.Join(importReport.Missing, ", ")))
		} else if importReport.Game == nil {
			report.Problems = append(report.Problems, "core pack cannot be imported")
		} else {
			report.Title = game.Title
		}
	}

	sort.Strings(report.Problems)
	sort.Strings(report.Warnings)
	return report, nil
}

func normalizeScaffoldOptions(absRoot string, opts ScaffoldOptions) ScaffoldOptions {
	opts.Title = strings.TrimSpace(opts.Title)
	if opts.Title == "" {
		base := strings.TrimSpace(filepath.Base(absRoot))
		if base == "" || base == "." || base == string(filepath.Separator) {
			base = defaultStoryTitle
		}
		opts.Title = base
	}

	opts.InitialScene = strings.TrimSpace(opts.InitialScene)
	if opts.InitialScene == "" {
		opts.InitialScene = defaultInitialScene
	}
	return opts
}

func ensureRootDir(absRoot string) error {
	info, err := os.Stat(absRoot)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("target path is not a directory: %s", absRoot)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat target path: %w", err)
	}
	if err := os.MkdirAll(absRoot, 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}
	return nil
}

func existingScaffoldFiles(absRoot string) ([]string, error) {
	var conflicts []string
	for _, file := range scaffoldFiles {
		target := filepath.Join(absRoot, filepath.FromSlash(file.Path))
		if _, err := os.Stat(target); err == nil {
			conflicts = append(conflicts, file.Path)
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("stat %s: %w", file.Path, err)
		}
	}
	return conflicts, nil
}

func writeFileAtomic(target string, content []byte, perm fs.FileMode, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("file already exists: %s", target)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat target file: %w", err)
		}
	}

	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	tempFile, err := os.CreateTemp(dir, "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if _, err := tempFile.Write(content); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tempFile.Chmod(perm); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if overwrite {
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove existing target: %w", err)
		}
	}
	if err := os.Rename(tempPath, target); err != nil {
		return fmt.Errorf("replace target file: %w", err)
	}
	return nil
}

func collectPackFiles(root string) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(current)
		if err != nil {
			return err
		}
		files[normalizePackPath(relative)] = string(content)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read pack files: %w", err)
	}
	return files, nil
}

func validateMetadata(raw string, report *ValidationReport) {
	var metadata roleplay.GameMetadata
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		report.Problems = append(report.Problems, fmt.Sprintf("metadata.json parse failed: %v", err))
		return
	}

	title := metadata.GameTitle()
	if title == "" {
		report.Problems = append(report.Problems, "metadata.json missing title")
	} else {
		report.Title = title
	}
	if strings.TrimSpace(metadata.InitialScene) == "" {
		report.Problems = append(report.Problems, "metadata.json missing initialScene")
	}
	if len(metadata.ScenePositions) == 0 {
		report.Warnings = append(report.Warnings, "metadata.json scenePositions is empty; the map panel will not show markers")
	}
	for sceneID, pos := range metadata.ScenePositions {
		if strings.TrimSpace(sceneID) == "" {
			report.Problems = append(report.Problems, "metadata.json scenePositions contains an empty scene id")
			continue
		}
		if len(pos) < 2 {
			report.Problems = append(report.Problems, fmt.Sprintf("metadata.json scenePositions.%s must contain [x, y]", sceneID))
			continue
		}
		if pos[0] < 0 || pos[0] > 100 || pos[1] < 0 || pos[1] > 100 {
			report.Warnings = append(report.Warnings, fmt.Sprintf("metadata.json scenePositions.%s is outside 0-100 map bounds", sceneID))
		}
	}
}

func validateBriefing(raw string, report *ValidationReport) {
	var briefing briefingTemplate
	if err := json.Unmarshal([]byte(raw), &briefing); err != nil {
		report.Problems = append(report.Problems, fmt.Sprintf("briefing.json parse failed: %v", err))
		return
	}
	if strings.TrimSpace(briefing.Title) == "" {
		report.Problems = append(report.Problems, "briefing.json missing title")
	}
	if len(briefing.Items) == 0 {
		report.Warnings = append(report.Warnings, "briefing.json items is empty; players will see no opening rules")
	}
	for index, item := range briefing.Items {
		if strings.TrimSpace(item.ID) == "" {
			report.Problems = append(report.Problems, fmt.Sprintf("briefing.json items[%d] missing id", index))
		}
		if strings.TrimSpace(item.Text) == "" {
			report.Problems = append(report.Problems, fmt.Sprintf("briefing.json items[%d] missing text", index))
		}
	}
	if strings.TrimSpace(briefing.ConfirmText) == "" {
		report.Warnings = append(report.Warnings, "briefing.json missing confirmText")
	}
}

func validatePhotoMetadata(raw string, report *ValidationReport) {
	var mapping map[string]string
	if err := json.Unmarshal([]byte(raw), &mapping); err != nil {
		report.Problems = append(report.Problems, fmt.Sprintf("photo/metadata.json parse failed: %v", err))
		return
	}
	if len(mapping) == 0 {
		report.Warnings = append(report.Warnings, "photo/metadata.json is empty; scene images are disabled")
	}
	for sceneID, fileName := range mapping {
		if strings.TrimSpace(sceneID) == "" {
			report.Problems = append(report.Problems, "photo/metadata.json contains an empty scene id")
		}
		if err := validateAssetFileName(fileName, supportedImageExts()); err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("photo/metadata.json %s: %v", sceneID, err))
		}
	}
}

func validateBGMMetadata(raw string, report *ValidationReport) {
	var metadata bgmMetadataTemplate
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		report.Problems = append(report.Problems, fmt.Sprintf("bgm/metadata.json parse failed: %v", err))
		return
	}
	if len(metadata.Tracks) == 0 {
		report.Warnings = append(report.Warnings, "bgm/metadata.json tracks is empty; BGM is disabled")
	}

	trackIDs := make(map[string]bool, len(metadata.Tracks))
	for id, track := range metadata.Tracks {
		if strings.TrimSpace(id) == "" {
			report.Problems = append(report.Problems, "bgm/metadata.json contains an empty track id")
			continue
		}
		trackIDs[id] = true
		if strings.TrimSpace(track.Name) == "" {
			report.Warnings = append(report.Warnings, fmt.Sprintf("bgm/metadata.json track %s missing name", id))
		}
		if err := validateAssetFileName(track.File, supportedAudioExts()); err != nil {
			report.Problems = append(report.Problems, fmt.Sprintf("bgm/metadata.json track %s: %v", id, err))
		}
	}
	for sceneID, trackID := range metadata.SceneDefaults {
		if strings.TrimSpace(sceneID) == "" {
			report.Problems = append(report.Problems, "bgm/metadata.json sceneDefaults contains an empty scene id")
		}
		if strings.TrimSpace(trackID) == "" {
			report.Problems = append(report.Problems, fmt.Sprintf("bgm/metadata.json sceneDefaults.%s missing track id", sceneID))
			continue
		}
		if !trackIDs[trackID] {
			report.Problems = append(report.Problems, fmt.Sprintf("bgm/metadata.json sceneDefaults.%s references unknown track %s", sceneID, trackID))
		}
	}
}

func validateAssetFileName(fileName string, supported []string) error {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return errors.New("missing file")
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	cleaned := path.Clean(normalized)
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") || strings.Contains(cleaned, "/../") {
		return fmt.Errorf("invalid relative file path %q", fileName)
	}
	ext := strings.ToLower(path.Ext(cleaned))
	if !slices.Contains(supported, ext) {
		return fmt.Errorf("unsupported file extension %q", ext)
	}
	return nil
}

func supportedImageExts() []string {
	return []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp"}
}

func supportedAudioExts() []string {
	return []string{".mp3", ".ogg", ".wav", ".m4a", ".webm"}
}

func normalizePackPath(name string) string {
	value := strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	value = strings.TrimPrefix(value, "./")
	value = strings.TrimPrefix(value, "/")
	value = path.Clean(value)
	value = strings.TrimPrefix(value, "./")
	value = strings.TrimPrefix(value, "/")
	return strings.ToLower(value)
}

func metadataJSON(opts ScaffoldOptions) string {
	positions := map[string][]float64{
		"entrance":    {85, 80},
		"living_room": {40, 55},
		"bedroom":     {12.5, 70},
		"bathroom":    {12.5, 55},
		"kitchen":     {80, 25},
	}
	if _, ok := positions[opts.InitialScene]; !ok {
		positions[opts.InitialScene] = []float64{85, 80}
	}
	return mustPrettyJSON(metadataTemplate{
		Title:          opts.Title,
		InitialScene:   opts.InitialScene,
		ScenePositions: positions,
	})
}

func briefingJSON(_ ScaffoldOptions) string {
	return mustPrettyJSON(briefingTemplate{
		Title:       "你需要记住的规则",
		Description: "在游戏开始前，请先确认这些规则。",
		Items: []briefingItem{
			{ID: "keep-watch", Text: "留意时间和环境变化"},
			{ID: "ask-help", Text: "遇到不对劲的事情时优先求助"},
			{ID: "follow-rules", Text: "不要忽略已经知道的规则"},
		},
		ConfirmText: "我已记住",
	})
}

func achievementsJSON(_ ScaffoldOptions) string {
	return mustPrettyJSON(struct {
		Achievements []roleplay.AchievementDefinition `json:"achievements"`
	}{
		Achievements: []roleplay.AchievementDefinition{
			{
				ID:      "truth_seeker",
				Title:   "真相追索者",
				Type:    roleplay.AchievementTypeAITriggered,
				Trigger: "玩家通过调查关键线索并主动说出隐藏真相时触发。",
				Ending: roleplay.Ending{
					ID:    "truth_revealed",
					Title: "真相追索者",
					Kind:  "good",
				},
			},
			{
				ID:    "one_life_clear",
				Title: "一次通关",
				Type:  roleplay.AchievementTypeRuleBased,
				Ending: roleplay.Ending{
					ID:    "one_life_clear",
					Title: "一次通关",
					Kind:  "good",
				},
				Rule: &roleplay.AchievementRule{
					Kind:               roleplay.AchievementRuleOneLife,
					EndingKind:         "good",
					ForbidSnapshotFork: true,
				},
			},
		},
	})
}

func photoMetadataJSON(opts ScaffoldOptions) string {
	mapping := map[string]string{
		"entrance":    "entrance.png",
		"living_room": "living_room.png",
		"bedroom":     "bedroom.png",
		"bathroom":    "bathroom.png",
		"kitchen":     "kitchen.png",
		"game_over":   "game_over.png",
	}
	if _, ok := mapping[opts.InitialScene]; !ok {
		mapping[opts.InitialScene] = safeAssetBase(opts.InitialScene) + ".png"
	}
	return mustPrettyJSON(mapping)
}

func bgmMetadataJSON(opts ScaffoldOptions) string {
	return mustPrettyJSON(bgmMetadataTemplate{
		Tracks: map[string]bgmTrackTemplate{
			"daily":   {Name: "日常氛围", File: "daily.mp3"},
			"tension": {Name: "紧张氛围", File: "tension.mp3"},
			"bad_end": {Name: "失败结局", File: "bad_end.mp3"},
		},
		SceneDefaults: map[string]string{
			opts.InitialScene: "daily",
			"kitchen":         "tension",
		},
	})
}

func sceneMarkdown(opts ScaffoldOptions) string {
	return fmt.Sprintf(`# 开场场景

## 场景 ID
- %s

## 当前情境
- 时间：
- 地点：
- 主角状态：
- 开场异常：

## 第一幕目标
- 让用户理解自己在哪里、现在有什么危险信号。
- 给出 2-4 个自然的行动方向，不要直接揭露隐藏真相。
`, opts.InitialScene)
}

func ruleMarkdown(_ ScaffoldOptions) string {
	return `# 剧情规则

## 用户开局前已知规则
见 briefing.json。这些规则已经在开局前展示给用户，AI 需要默认用户知道并可以遵守，不要把它们当作隐藏真相。

## 隐藏规则
- 这里写用户不应提前知道，但 AI 推进剧情时必须遵守的规则。
- 如果用户违反关键规则，需要进入对应失败、循环或中立结局。

## 环境变量
- 时间：
- 天气：
- 固定事件：

## 阶段规则

### 第一阶段
- 约束
- 行为

### 第二阶段
- 约束
- 行为
`
}

func trueMarkdown(_ ScaffoldOptions) string {
	return `# 隐藏真相

- 这里写故事真正发生了什么。
- AI 可以用这些信息做推理，但不能直接告诉用户。
- 真相需要通过场景线索、规则后果和结局逐步体现。
`
}

func memoryMarkdown(_ ScaffoldOptions) string {
	return `# 记事本

## 初始状态
- 当前阶段：
- 用户已知信息：
- 关键资源：

## 行动记录
- 尚未开始。

## 分支记录
- 尚未进入分支。
`
}

func endingsMarkdown(_ ScaffoldOptions) string {
	return `# 结局列表

## 好结局 1
- 触发条件：
- 结局描述：

## 好结局 2
- 触发条件：
- 结局描述：

## 坏结局 1
- 触发条件：
- 结局描述：

## 坏结局 2
- 触发条件：
- 结局描述：

## 循环结局
- 触发条件：
- 结局描述：
`
}

func mustPrettyJSON(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data) + "\n"
}

func safeAssetBase(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "scene"
	}
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	value = replacer.Replace(value)
	if value == "." || value == ".." {
		return "scene"
	}
	return value
}
