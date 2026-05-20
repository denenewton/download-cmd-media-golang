package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// desenhaBarra constrói a string visual da barra de progresso baseado no valor de 0 a 100
func desenhaBarra(percentual float64, velocidade string) {
	const larguraBarra = 20
	posicaoPreenchida := int((percentual / 100.0) * larguraBarra)

	barra := "["
	for i := 0; i < larguraBarra; i++ {
		if i < posicaoPreenchida {
			barra += "█"
		} else {
			barra += "░"
		}
	}
	barra += "]"

	// \r limpa a linha e atualiza a barra, porcentagem e velocidade no mesmo lugar
	fmt.Printf("\r📥 Progresso: %s %.1f%% | Velocidade: %s", barra, percentual, velocidade)
}

// DownloadByQuery aceita busca/URL, formato e o caminho de destino onde o arquivo deve ser salvo
func DownloadByQuery(input string, apenasAudio bool, pastaDestino string) error {
	var arguments []string

	if strings.Contains(input, "http") {
		arguments = append(arguments, input)
	} else {
		arguments = append(arguments, "ytsearch1:"+input)
	}

	caminhoSaida := filepath.Join(pastaDestino, "%(title)s.%(ext)s")

	// Força o yt-dlp a enviar atualizações em novas linhas constantemente
	arguments = append(arguments, "--newline")

	if apenasAudio {
		arguments = append(arguments,
			"-x",
			"--audio-format", "mp3",
			"--extractor-args", "youtube:player_client=android,web",
			"-o", caminhoSaida,
		)
	} else {
		arguments = append(arguments,
			"--extractor-args", "youtube:player_client=android,web",
			"-o", caminhoSaida,
		)
	}

	cmd := exec.Command("yt-dlp", arguments...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	// Expressão regular para capturar com precisão a porcentagem e a velocidade da string do yt-dlp
	re := regexp.MustCompile(`\[download\]\s+(\d+\.\d+)%\s+of\s+\S+\s+at\s+(\S+)`)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		linha := scanner.Text()

		if strings.HasPrefix(linha, "[download]") && strings.Contains(linha, "%") {
			matches := re.FindStringSubmatch(linha)
			if len(matches) == 3 {
				percentual, err := strconv.ParseFloat(matches[1], 64)
				velocidade := matches[2]
				if err == nil {
					desenhaBarra(percentual, velocidade)
				}
			}
		} else if strings.HasPrefix(linha, "[ExtractAudio]") {
			fmt.Print("\r🎵 Convertendo arquivo para MP3, por favor aguarde...               ")
		}
	}

	fmt.Println()
	return cmd.Wait()
}

func main() {
	var busca string
	var opcao int

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Erro ao localizar pasta do usuário: %v\n", err)
		return
	}

	pastaMusicas := filepath.Join(homeDir, "Music")
	pastaVideos := filepath.Join(homeDir, "Videos")

	fmt.Print("Digite o link ou o nome da música/vídeo: ")
	reader := bufio.NewReader(os.Stdin)
	busca, _ = reader.ReadString('\n')
	busca = strings.TrimSpace(busca)

	fmt.Println("\nEscolha o formato:")
	fmt.Println("1 - Vídeo (Salvar em ~/Videos)")
	fmt.Println("2 - Apenas Música (Salvar em ~/Music)")
	fmt.Print("Opção: ")
	fmt.Scanln(&opcao)

	var pastaDestino string
	apenasAudio := (opcao == 2)

	if apenasAudio {
		pastaDestino = pastaMusicas
	} else {
		pastaDestino = pastaVideos
	}

	os.MkdirAll(pastaDestino, os.ModePerm)

	fmt.Println("\n🔍 Buscando no YouTube...")
	err = DownloadByQuery(busca, apenasAudio, pastaDestino)
	if err != nil {
		fmt.Printf("\n❌ Erro no processo: %v\n", err)
		return
	}

	fmt.Println("✅ Download concluído com sucesso!")
}
