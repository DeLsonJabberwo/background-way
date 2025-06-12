#cd $HOME/.local/bin
cd /usr/bin
sudo curl -LO https://github.com/DaltonJabberwo/background-way/releases/latest/download/background
sudo chmod +x background

mkdir $HOME/.background
cd $HOME/.background
curl -LO https://github.com/DaltonJabberwo/background-way/releases/latest/download/conf.json

cd ~
echo "$HOME/.background/conf.json:"
cat $HOME/.background/conf.json

