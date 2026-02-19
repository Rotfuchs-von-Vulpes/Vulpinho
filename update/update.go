package update

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"vulpinho/log"
)

var logger *slog.Logger

func DownloadFile(source string) error {
	resp, err := http.Get(source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad status code: %s", resp.Status)
	}

	out, err := os.Create("commands/wh40k/data/" + filepath.Base(source))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func UpdateWarhammer() {
	logger = log.Logger
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

	for _, source := range sources {
		if err := DownloadFile(source); err != nil {
			logger.Error("Erro ao atualizar dados de Warhammer 40k", "error", err)
			break
		}
		logger.Info("Arquivo carregado com sucesso", "file", filepath.Base(source))
	}

	logger.Info("Dados sobre Warhammer atualizado com sucesso")
}
