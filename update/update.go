package update

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func downloadFile(url string) error {
	fileName := filepath.Base(url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad status code: %s", resp.Status)
	}

	out, err := os.Create("resources/wh40k/" + fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	if err == nil {
		slog.Info("Arquivo carregado com sucesso", "file", fileName)
		return nil
	}

	return err
}

func UpdateWarhammer() {
	sources := []string{
		"http://wahapedia.ru/wh40k10ed/Abilities.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_abilities.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_detachment_abilities.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_enhancements.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_keywords.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_leader.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_models_cost.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_models.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_options.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_stratagems.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_unit_composition.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets_wargear.csv",
		"http://wahapedia.ru/wh40k10ed/Datasheets.csv",
		"http://wahapedia.ru/wh40k10ed/Detachments.csv",
		"http://wahapedia.ru/wh40k10ed/Detachment_abilities.csv",
		"http://wahapedia.ru/wh40k10ed/Factions.csv",
		"http://wahapedia.ru/wh40k10ed/Stratagems.csv",
		"http://wahapedia.ru/wh40k10ed/Enhancements.csv",
	}

	errOcurr := false
	for _, source := range sources {
		if err := downloadFile(source); err != nil {
			slog.Error("Erro ao atualizar dados de Warhammer 40k", "error", err)
			errOcurr = true
			continue
		}
	}

	if !errOcurr {
		slog.Info("Dados sobre Warhammer atualizado com sucesso")
	}
}

func GetLastEdit() {
	if err := downloadFile("http://wahapedia.ru/wh40k10ed/Last_update.csv"); err != nil {
		slog.Error("Erro ao fazer o download da data da ultima edição.", "error", err)
		return
	}

	file, err := os.Open("resources/wh40k/Last_update.csv")
	if err != nil {
		slog.Error("Erro ao abrir arquivo de ultima edição.", "err", err)
		return
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.Comma = '|'
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		slog.Error("Incapaz de parsear CSV.", "error", err.Error())
		return
	}

	records[0][0], _ = strings.CutPrefix(records[0][0], string(rune(65279)))
	lastUpEdit := records[1][0]

	dummyFile, err := os.OpenFile("resources/wh40k/dummy.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("Incapaz de abrir arquivo da data do ultimo download.", "err", err)
		return
	}
	defer dummyFile.Close()

	var dummyDate string
	if content, err := io.ReadAll(dummyFile); err == nil {
		dummyDate = string(content)
	} else {
		slog.Error("Incapaz de ler arquivo.", "err", err)
		return
	}

	if dummyDate != lastUpEdit {
		slog.Info("Atualizando dados sobre Warhammer")
		UpdateWarhammer()
		dummyFile.Truncate(0)
		dummyFile.Seek(0, 0)
		dummyFile.WriteString(lastUpEdit)
	}
}
